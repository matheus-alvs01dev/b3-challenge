package config

import (
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrap(err, "viper read")
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return errors.Wrap(err, "viper unmarshal")
	}

	return nil
}

var cfg *Config //nolint:gochecknoglobals

type Config struct {
	APIPort            uint16 `mapstructure:"API_PORT"`
	DatabaseDSN        string `mapstructure:"DB_DSN"`
	ParserWorkersCount int    `mapstructure:"PARSER_WORKER_COUNT"`
	DBWorkersCount     int    `mapstructure:"DB_WORKER_COUNT"`
	BatchSize          int    `mapstructure:"BATCH_SIZE"`
}

func GetAPIPort() uint16 {
	return cfg.APIPort
}

func GetDatabaseDSN() string {
	return cfg.DatabaseDSN
}

func GetParserWorkersCount() int {
	if cfg.ParserWorkersCount == 0 {
		return runtime.NumCPU()
	}

	return cfg.ParserWorkersCount
}

func GetBatchSize() int {
	const defaultBatchSize = 50000

	if cfg.BatchSize == 0 {
		return defaultBatchSize
	}

	return cfg.BatchSize
}

func GetDBWorkersCount() int {
	if cfg.DBWorkersCount == 0 {
		return runtime.NumCPU()
	}

	return cfg.DBWorkersCount
}
