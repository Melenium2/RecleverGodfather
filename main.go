package main

import (
	"RecleverGodfather/grandlog"
	"RecleverGodfather/grandlog/internallogger"
	"RecleverGodfather/grandlog/loggerepo"
	"RecleverGodfather/handlers"
	"RecleverGodfather/remoteclients"
	"fmt"
	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd/consul"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	"github.com/jmoiron/sqlx"
	_ "github.com/mailru/go-clickhouse"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		httpAddr     = os.Getenv("http_port")
		_            = os.Getenv("grpc_port")
		consulAddr   = os.Getenv("consul_addr")
		tgToken      = os.Getenv("tg_token")
		tgChatId     = os.Getenv("tg_chat_id")
		loggerSource = os.Getenv("logger_db")
		configDir    = os.Getenv("config_dir")
	)
	if httpAddr == "" || consulAddr == "" || loggerSource == "" {
		log.Fatal("Error in main. Need to provide environment vars first")
	}
	if tgToken == "" || tgChatId == "" {
		log.Fatal("You need to provide telegram token and chat id")
	}
	if configDir == "" {
		configDir = "."
	}
	chatId, _ := strconv.Atoi(tgChatId)
	var telegramBot = internallogger.NewTelegramLogger(tgToken, chatId)
	var logger = createLogger(createLoggerDb(loggerSource, configDir), telegramBot)

	var consulClient consul.Client
	{
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulAddr
		cl, err := api.NewClient(consulConfig)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}

		consulClient = consul.NewClient(cl)
	}

	r := mux.NewRouter()
	r.PathPrefix("/guard").Handler(http.StripPrefix("/guard", remoteclients.NewGuardClient(consulClient, logger)))
	r.PathPrefix("/recr").Handler(http.StripPrefix("/recr", remoteclients.NewRecruiterClient(consulClient, logger)))
	r.PathPrefix("/right").Handler(http.StripPrefix("/right", remoteclients.NewRightHandClient(consulClient, logger)))
	r.HandleFunc("/log", handlers.Log(logger))
	printRouter(logger, r)

	log.Print("look for changes...!!!2d")

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "http", "addr", httpAddr)
		errs <- http.ListenAndServe(httpAddr, r)
	}()

	go func() {
		up := telegramBot.ServeUpdates(internallogger.GenerateUpdateConfig(0))
		defer telegramBot.CloseUpdates()
		defer up.Clear()
		logger.Log("type", "[Info]", "service", "godfather", "action", "start serving updates from telegram")
		for {
			update, ok := <-up
			if !ok {
				return
			}
			telegramBot.Sendlog(update.Message.Chat.Id, "Message accepted")
		}
	}()

	// log errs
	logger.Log("Terminate", <-errs)
}

func createLoggerDb(dbURL, configDir string) *sqlx.DB {
	log.Print("Connect to db url ", dbURL, " ...")
	c, err := sqlx.Connect("clickhouse", dbURL)
	if err != nil {
		log.Print(err)
		time.Sleep(time.Second * 15)
		createLoggerDb(dbURL, configDir)
	}
	log.Print("Connected to db.", " Init schema...")

	ddl, err := ioutil.ReadFile(fmt.Sprintf("%s/config/schema.sql", configDir))
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Read schema from file...")
	if _, err := c.Exec(string(ddl)); err != nil {
		if strings.Contains(err.Error(), "Code: 57") {
			newddl := strings.ReplaceAll(string(ddl), "create table if not exists", "ATTACH TABLE")
			if _, err := c.Exec(newddl); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	log.Print("Schema created.")

	return c
}

func createLogger(loggerDb *sqlx.DB, telegramBot internallogger.InternalLogger) grandlog.GrandLogger {
	var logger grandlog.GrandLogger
	{
		var defaultLogger kitlog.Logger
		{
			defaultLogger = kitlog.NewLogfmtLogger(os.Stderr)
			defaultLogger = kitlog.With(defaultLogger, "ts", kitlog.DefaultTimestampUTC)
			defaultLogger = kitlog.With(defaultLogger, "caller", kitlog.DefaultCaller)
		}
		loggerRepo := loggerepo.NewClickhouseLogger(loggerDb, defaultLogger)
		defaultLogger.Log("[Info]", "Logger db initialized")
		defaultLogger.Log("[Info]", "Internal logger initialized")

		logger = grandlog.NewGrandLogger(loggerRepo, defaultLogger, telegramBot)
	}
	logger.Log("[Info]", "logger created")
	return logger
}

func printRouter(logger grandlog.GrandLogger, router *mux.Router) {
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		temp, err := route.GetPathTemplate()
		if err != nil {
			logger.Log("type", "[Error]", "service", "godfather", "trace", err)
			return err
		}

		logger.Log("type", "[Info]", "service", "godfather", "route", temp)
		return nil
	})
}
