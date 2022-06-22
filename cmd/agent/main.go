package main

import (
	"flag"
	"log"
	"time"

	"github.com/MaximkaSha/log_tools/internal/agent"
	"github.com/caarlos0/env/v6"
)

var (
	srvAdressArg      *string
	reportIntervalArg *time.Duration
	pollIntervalArg   *time.Duration
	keyFile           *string
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	reportIntervalArg = flag.Duration("r", time.Duration(10*time.Second), "report to server interval in seconds (default 10s)")
	pollIntervalArg = flag.Duration("p", time.Duration(2*time.Second), "poll interval in seconds (default 2s)")
	keyFile = flag.String("k", "key.txt", "path to key file (default key.txt)")
}

func main() {
	var cfg agent.Config
	var envCfg = make(map[string]bool)
	opts := env.Options{
		OnSet: func(tag string, value interface{}, isDefault bool) {
			envCfg[tag] = isDefault
		},
	}
	err := env.Parse(&cfg, opts)

	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	var a = flag.Lookup("a")
	if envCfg["ADDRESS"] && a != nil {
		cfg.Server = *srvAdressArg
	}
	a = flag.Lookup("r")
	if envCfg["REPORT_INTERVAL"] && a != nil {
		cfg.ReportInterval = time.Duration(*reportIntervalArg)
	}
	a = flag.Lookup("p")
	if envCfg["POLL_INTERVAL"] && a != nil {
		cfg.PollInterval = time.Duration(*pollIntervalArg)
	}
	a = flag.Lookup("k")
	if envCfg["KEY"] && a != nil {
		cfg.KeyFile = *keyFile
	}
	agentService := agent.NewAgent(cfg)
	agentService.StartService()

}
