package main

import (
	"RecleverGodfather/config"
	"RecleverGodfather/grandlog"
	"RecleverGodfather/grandlog/internallogger"
	"RecleverGodfather/grandlog/loggerepo"
	"RecleverGodfather/handlers"
	"RecleverGodfather/remoteclients"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
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
	"syscall"
)

func main() {
	conf := createConfig()
	var (
		httpAddr    = conf.HTTPPort
		_           = conf.GRPCPort
		consulAddr  = conf.ConsulPort
		tgToken     = conf.TgToken
		tgChatId    = conf.TgChatId
		telegramBot = internallogger.NewTelegramLogger(tgToken, tgChatId)
		logger      = createLogger(createLoggerDb(conf.LoggerDBUrl), telegramBot)
	)
	var consulClient consul.Client
	{
		consulConfig := api.DefaultConfig()
		if len(consulAddr) > 0 {
			consulConfig.Address = consulAddr
		} else {
			logger.Log("Consul port is empty")
			os.Exit(1)
		}
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
	r.HandleFunc("/log", handlers.Log(logger))
	printRouter(logger, r)

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

func createConfig() *config.Config {
	var configPath string
	flag.StringVar(&configPath, "config-path", "config/config.toml", "path to config file")
	flag.Parse()
	c := config.NewConfig()
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func createLoggerDb(dbURL string) *sqlx.DB {
	log.Print("Connect to db url ", dbURL, " ...")
	c, err := sqlx.Connect("clickhouse", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to db.", "Init schema...")

	ddl, err := ioutil.ReadFile("config/schema.sql")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Read schema from file...")
	if _, err := c.Exec(string(ddl)); err != nil {
		log.Fatal(err)
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

		logger.Log("type", "[Error]", "service", "godfather", "route", temp)
		return nil
	})
}

