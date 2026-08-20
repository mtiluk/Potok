package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr        string
	DataDir     string
	DatabaseURL string
}

func LoadConfig() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	cfg := Config{
		Addr:        os.Getenv("ADDR"),
		DataDir:     os.Getenv("DATA_DIR"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if err := cfg.ValidateConfig(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) ValidateConfig() error {
	if cfg.Addr == "" {
		return errors.New("ADDR is required")
	}
	if cfg.DataDir == "" {
		return errors.New("DATA_DIR is required")
	}
	if cfg.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	return nil
}
