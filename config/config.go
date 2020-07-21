package config

import (
	"github.com/caarlos0/env"
	zl "github.com/rs/zerolog/log"
)

type Config struct {
	Listen         string `env:"LISTEN" envDefault:"localhost:8081"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"debug"`
	CreateDatabase bool   `env:"CREATE_DB" envDefault:"false"`
	PgSQL          string `env:"PGSQL" envDefault:"postgres://postgres:1234@localhost:5432/postgres?sslmode=disable"`
}

// LoadConfig return struct config
func LoadConfig() Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		zl.Fatal().Err(err).
			Msg("Can't parse env args")
	}
	return cfg
}
