package config

import (
	"flag"
	"fmt"
	"slices"
	"strings"
)

type Config struct {
	Workers     int
	LoggerLevel string
	Host        string
	Port        int
}

func NewConfig() (*Config, error) {
	cfg := new(Config)
	cfg.getConfigParams()

	if err := cfg.validateLogLevel(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) getConfigParams() {
	flag.IntVar(&c.Workers, "workers", 1, "Number of workers to spawn")
	flag.StringVar(&c.LoggerLevel, "logger-level", "local", "Logger level")
	flag.StringVar(&c.Host, "host", "0.0.0.0", "HTTP Host")
	flag.IntVar(&c.Port, "port", 8080, "HTTP Port")
	flag.Parse()
}

func (c *Config) validateLogLevel() error {
	availableLevels := []string{"local", "dev", "prod"}

	if !slices.Contains(availableLevels, c.LoggerLevel) {
		return fmt.Errorf(
			"logger level '%s' is not supported, avaliable levels: %s",
			c.LoggerLevel,
			strings.Join(availableLevels, ", "),
		)
	}

	return nil
}
