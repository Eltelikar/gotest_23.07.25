package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string      `yaml:"env" env:"ENV" env-default:"local" env-requered:"true"`
	HTTPServer *HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:"localhost"`
	Port         string        `yaml:"port" env-default:"8080"`
	Timeout      time.Duration `yaml:"timeout" env-default:"10s"`
	Idle_timeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIGPATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("cannot file configfile by path: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}

	return &cfg
}
