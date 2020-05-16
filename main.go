package main

import (
	"RecleverGodfather/config"
	"RecleverGodfather/grandlog"
	"RecleverGodfather/grandlog/internallogger"
	"RecleverGodfather/grandlog/loggerepo"
	"RecleverGodfather/remoteclients"
	"encoding/json"
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
		httpAddr   = conf.HTTPPort
		_          = conf.GRPCPort
		consulAddr = conf.ConsulPort
		logger     = createLogger(createLoggerDb(conf.LoggerDBUrl))
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
	r.PathPrefix("/v1").Handler(http.StripPrefix("/v1", remoteclients.NewGuardClient(consulClient, logger)))
	r.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		log := &loggerepo.SingleLog{}
		if err := json.NewDecoder(r.Body).Decode(log); err != nil {
			logger.Log("[Error]", err)
			return
		}
		if log.MessageType == "" {
			logger.Log("[Error]", "empty request")
			return
		}
		if err := logger.LogObject(r.Context(), log); err != nil {
			logger.Log("[Error]", err)
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("Complete"))
	})

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

func createLogger(loggerDb *sqlx.DB) grandlog.GrandLogger {
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
		internalLogger := internallogger.NewTelegramLogger()
		defaultLogger.Log("[Info]", "Internal logger initialized")

		logger = grandlog.NewGrandLogger(loggerRepo, defaultLogger, internalLogger)
	}
	logger.Log("[Info]", "logger created")
	return logger
}

