package main

import (
	"flag"
	"fmt"
	"time"
)

type Config struct {
	From    int
	To      int
	Workers int
	Timeout time.Duration
}

func parseConfig() (Config, error) {
	var cfg Config
	flag.IntVar(&cfg.From, "from", -1, "flag from")
	flag.IntVar(&cfg.To, "to", -1, "flag to")
	flag.IntVar(&cfg.Workers, "workers", 10, "number of concurrent workers")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "timeout for each HTTP request")

	flag.Parse()

	if cfg.From <= 0 {
		return Config{}, fmt.Errorf("--from must be greater than 0, got %d", cfg.From)
	}
	if cfg.To <= 0 {
		return Config{}, fmt.Errorf("--to must be greater than 0, got %d", cfg.To)
	}
	if cfg.From > cfg.To {
		return Config{}, fmt.Errorf("--from must not be greater than --to: from=%d, to=%d", cfg.From, cfg.To)
	}
	if cfg.Workers <= 0 {
		return Config{}, fmt.Errorf("--workers must be greater than 0, got %d", cfg.Workers)
	}
	if cfg.Timeout <= 0 {
		return Config{}, fmt.Errorf("--timeout must be greater than 0, got %v", cfg.Timeout)
	}

	return cfg, nil
}
