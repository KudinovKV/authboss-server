package main

import (
	"os"

	"github.com/KudinovKV/authboss-server/web"

	"github.com/KudinovKV/authboss-server/config"
	"github.com/KudinovKV/authboss-server/database"
	"github.com/rs/zerolog"
	zl "github.com/rs/zerolog/log"
)

func initLogLevel(LogLevel string) {
	logLevel, err := zerolog.ParseLevel(LogLevel)
	if err != nil {
		zl.Fatal().Err(err).Msgf("Can't parse loglevel")
	}
	zerolog.SetGlobalLevel(logLevel)
	zl.Logger = zl.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	cfg := config.LoadConfig()
	initLogLevel(cfg.LogLevel)
	db, err := database.InitDb(cfg.PgSQL, cfg.CreateDatabase)
	if err != nil {
		zl.Fatal().Err(err).Msg("Can't init database")
	}
	defer func() {
		err := db.Close()
		if err != nil {
			zl.Warn().Err(err).Msg("Can't close database")
		}
	}()
	web.InitServer(cfg.Listen, db)
}
