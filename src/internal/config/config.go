package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DBUser           string `env:"DB_USER" env-default:"postgres"`
	DBPassword       string `env:"DB_PASSWORD" env-required:"true"`
	DBHost           string `env:"DB_HOST" env-default:"localhost"`
	DBPort           string `env:"DB_PORT" env-default:"5432"`
	DBName           string `env:"DB_NAME" env-default:"tictactoe"`
	DBSSLMode        string `env:"DB_SSLMODE" env-default:"disable"`
	JWTAccessSecret  string `env:"JWT_ACCESS_SECRET" env-default:"fallback_access_secret_key_32_bytes"`
	JWTRefreshSecret string `env:"JWT_REFRESH_SECRET" env-default:"fallback_refresh_secret_key_32_bytes"`
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func NewConfig() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return &cfg, nil
}
