package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTP       HTTP       `yaml:"http"`
	GRPC       GRPC       `yaml:"grpc"`
	Prometheus Prometheus `yaml:"prometheus"`
	Database   Database   `yaml:"database"`
	JWT        JWT        `yaml:"jwt"`
}

type HTTP struct {
	Host        string        `yaml:"host" env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port        string        `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"30s"`
}

type GRPC struct {
	IsAble bool   `yaml:"is_able" env:"GRPC_IS_ABLE" env-default:"true"`
	Host   string `yaml:"host" env:"GRPC_HOST" env-default:"0.0.0.0"`
	Port   string `yaml:"port" env:"GRPC_PORT" env-default:"3000"`
}

type Prometheus struct {
	IsAble      bool          `yaml:"is_able" env:"PROMETHEUS_IS_ABLE" env-default:"true"`
	Host        string        `yaml:"host" env:"PROMETHEUS_HOST" env-default:"0.0.0.0"`
	Port        string        `yaml:"port" env:"PROMETHEUS_PORT" env-default:"3000"`
	Timeout     time.Duration `yaml:"timeout" env:"PROMETHEUS_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"PROMETHEUS_IDLE_TIMEOUT" env-default:"30s"`
}

type Database struct {
	Protocol string `yaml:"protocol" env:"DB_PROTOCOL" env-default:"postgres"`
	Host     string `yaml:"host" env:"DB_HOST" env-default:"db"`
	Port     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Name     string `yaml:"name" env:"DB_NAME" env-default:"pvz"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE" env-default:"disable"`
}

type JWT struct {
	SecretKey string        `yaml:"secret" env:"JWT_SECRET" env-default:"secret"`
	ExpiresIn time.Duration `yaml:"expires_in" env:"JWT_EXPIRES_IN" env-default:"24h"`
}

func Load() (*Config, error) {
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok {
		return nil, errors.New("CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("cannot read config: %s", err)
	}

	return &cfg, nil
}
