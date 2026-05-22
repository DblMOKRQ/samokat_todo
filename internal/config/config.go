package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

var errHTTPAddrEmpty = errors.New("HTTP_ADDR is empty")

const (
	defaultHTTPAddr            = "localhost:8080"
	defaultHTTPRequestTimeout  = 3 * time.Second
	defaultHTTPShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Addr            string
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

// Load загружает конфигурацию из переменных окружения
func Load() (Config, error) {
	cfg := Config{
		HTTP: HTTPConfig{
			Addr:            envString("HTTP_ADDR", defaultHTTPAddr),
			RequestTimeout:  envDuration("HTTP_REQUEST_TIMEOUT", defaultHTTPRequestTimeout),
			ShutdownTimeout: envDuration("HTTP_SHUTDOWN_TIMEOUT", defaultHTTPShutdownTimeout),
		},
	}

	if cfg.HTTP.Addr == "" {
		return Config{}, errHTTPAddrEmpty
	}

	return cfg, nil
}

func envString(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// envDuration читает duration из env
func envDuration(key string, def time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}

	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}

	if sec, err := strconv.Atoi(raw); err == nil && sec >= 0 {
		return time.Duration(sec) * time.Second
	}

	return def
}
