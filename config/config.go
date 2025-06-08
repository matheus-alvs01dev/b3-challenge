package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"runtime"
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

var cfg *Config

type Config struct {
	APIPort     uint16 `mapstructure:"API_PORT"`
	DatabaseDSN string `mapstructure:"DB_DSN"`
	WorkerCount int    `mapstructure:"WORKER_COUNT"`
	BatchSize   int    `mapstructure:"BATCH_SIZE"`
}

func GetAPIPort() uint16 {
	return cfg.APIPort
}

func GetDatabaseDSN() string {
	return cfg.DatabaseDSN
}

func GetWorkerCount() int {
	if cfg.WorkerCount == 0 {
		return runtime.NumCPU()
	}

	return cfg.WorkerCount
}

func GetBatchSize() int {
	if cfg.BatchSize == 0 {
		return 50000
	}

	return cfg.BatchSize
}
