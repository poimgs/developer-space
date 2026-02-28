package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port           string
	DatabaseURL    string
	FrontendURL    string
	SessionSecret  string
	ResendAPIKey   string
	ResendFromEmail string
	TelegramBotToken string
	TelegramChatID   string
	LogLevel       string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		SessionSecret:  os.Getenv("SESSION_SECRET"),
		ResendAPIKey:   os.Getenv("RESEND_API_KEY"),
		ResendFromEmail: os.Getenv("RESEND_FROM_EMAIL"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		LogLevel:    strings.ToLower(getEnv("LOG_LEVEL", "info")),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.SessionSecret == "" {
		return fmt.Errorf("SESSION_SECRET is required")
	}
	if c.ResendAPIKey == "" {
		return fmt.Errorf("RESEND_API_KEY is required")
	}
	if c.ResendFromEmail == "" {
		return fmt.Errorf("RESEND_FROM_EMAIL is required")
	}
	return nil
}

func (c *Config) IsSecure() bool {
	return strings.HasPrefix(c.FrontendURL, "https://")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
