package config

import (
	"log"

	env "github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	// PocketBase
	PBDev  bool   `env:"PB_DEV" envDefault:"true"`
	PBPort string `env:"PB_PORT" envDefault:"8090"`

	// Spotify
	SpotifyClientID     string `env:"SPOTIFY_CLIENT_ID"`
	SpotifyClientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
	SpotifyRedirectURI  string `env:"SPOTIFY_REDIRECT_URI" envDefault:"http://localhost:8090/auth/spotify/callback"`

	// Application
	AppEnv   string `env:"APP_ENV" envDefault:"dev"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`
}

// Load loads configuration from .env file and environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// MustLoad loads configuration and panics on error
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	return cfg
}

func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "dev"
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "prod"
}
