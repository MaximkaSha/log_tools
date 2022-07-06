package main

import (
	"flag"
	"log"
	"time"

	"github.com/MaximkaSha/log_tools/internal/agent"
	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg agent.Config
	var envCfg = make(map[string]bool)
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			envCfg[tag] = isDefault
		},
	}
	// Сначала читаем ключи кмд
	flag.StringVar(&cfg.Server, "a", "localhost:8080", "host:port (default localhost:8080)")
	flag.DurationVar(&cfg.ReportInterval, "r", time.Duration(10*time.Second), "report to server interval in seconds (default 10s)")
	flag.DurationVar(&cfg.PollInterval, "p", time.Duration(2*time.Second), "poll interval in seconds (default 2s)")
	flag.StringVar(&cfg.KeyFile, "k", "", "hmac key")
	flag.Parse()
	// Потом переписываем ключами из ENV, они имеют приоритет
	err := env.Parse(&cfg, opts)
	if err != nil {
		log.Fatal(err)
	}
	agentService := agent.NewAgent(cfg)
	agentService.StartService()

}
