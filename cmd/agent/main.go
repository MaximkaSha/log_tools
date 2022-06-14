package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/agent"
	"github.com/caarlos0/env/v6"
)

var (
	srvAdressArg      *string
	reportIntervalArg *time.Duration
	pollIntervalArg   *time.Duration
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	reportIntervalArg = flag.Duration("r", time.Duration(10*time.Second), "report to server interval in seconds (default 10s)")
	pollIntervalArg = flag.Duration("p", time.Duration(2*time.Second), "poll interval in seconds (default 2s)")
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
	agentService := agent.NewAgent()
	var pollInterval = cfg.PollInterval
	var reportInterval = cfg.ReportInterval
	//var logData = new(logData)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	//	var rtm runtime.MemStats
	log.Println("Logger start...")
	tickerCollect := time.NewTicker(pollInterval)
	tickerSend := time.NewTicker(reportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()
	for {
		select {
		case <-tickerCollect.C:
			agentService.CollectLogs()
		case <-tickerSend.C:
			agentService.SendLogsbyPost("http://" + cfg.Server + "/update/")
			agentService.SendLogsbyJSON("http://" + cfg.Server + "/update/")
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}

}
