package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Postgres     PostgresConfig `mapstructure:",squash"`
	ServerConfig ServerConfig   `mapstructure:",squash"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	Username string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASS"`
	Database string `mapstructure:"DB_DATABASE"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
}

type ServerConfig struct {
	Host              string        `mapstructure:"SERVER_HOST"`
	Port              string        `mapstructure:"SERVER_PORT"`
	ReadHeaderTimeout time.Duration `mapstructure:"READ_HEADER_TIMEOUT"`
	ReadTimeout       time.Duration `mapstructure:"READ_TIMEOUT"`
}

func Load() (Config, error) {
	return LoadFromFile(".env")
}

func LoadFromFile(path string) (Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("env")
	v.AutomaticEnv()

	setDefaults(v)
	if err := readConfig(v); err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("SERVER_HOST", "localhost")
	v.SetDefault("SERVER_PORT", "8080")
	v.SetDefault("READ_HEADER_TIMEOUT", "5s")
	v.SetDefault("READ_TIMEOUT", "5s")
}

func readConfig(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) || errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	return nil
}
