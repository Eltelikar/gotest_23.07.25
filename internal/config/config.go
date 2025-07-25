package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string       `yaml:"env" env:"ENV" env-default:"local" env-requered:"true"`
	StorageLink *StorageLink `yaml:"storage_link"`
	HTTPServer  *HTTPServer  `yaml:"http_server"`
}

type StorageLink struct {
	SQLDriver   string `yaml:"sql_driver" env-default:"postgres"`
	SQLUser     string `yaml:"sql_user" env-default:"postgres"`
	SQLPassword string `yaml:"sql_password" env-default:"postgres"`
	SQLHost     string `yaml:"sql_host" env-default:"0.0.0.0"`
	SQLPort     string `yaml:"sql_port" env-default:"8080"`
	SQLDBName   string `yaml:"sql_dbname" env-default:"wallet_app"`
	SQLSSLMode  string `yaml:"sql_sslmode" env-default:"disable"`
}

type HTTPServer struct {
	Address      string        `yaml:"address" env-default:":8080"`
	Timeout      time.Duration `yaml:"timeout" env-default:"10s"`
	Idle_timeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("cannot file configfile by path: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}

	return &cfg
}

func GetStorageLink(cfg *Config) string {
	storageLink := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.StorageLink.SQLDriver,
		cfg.StorageLink.SQLUser,
		cfg.StorageLink.SQLPassword,
		cfg.StorageLink.SQLHost,
		cfg.StorageLink.SQLPort,
		cfg.StorageLink.SQLDBName,
		cfg.StorageLink.SQLSSLMode,
	)

	if storageLink == "" {
		log.Fatalf("storage link is empty")
	}

	slog.Debug("Storage link set in config",
		slog.String("SQLDriver", cfg.StorageLink.SQLDriver),
		slog.String("SQLUser", cfg.StorageLink.SQLUser),
		slog.String("SQLPassword", cfg.StorageLink.SQLPassword),
		slog.String("SQLHost", cfg.StorageLink.SQLHost),
		slog.String("SQLPort", cfg.StorageLink.SQLPort),
		slog.String("SQLDBName", cfg.StorageLink.SQLDBName),
		slog.String("SQLSSLMode", cfg.StorageLink.SQLSSLMode),
	)
	return storageLink
}
