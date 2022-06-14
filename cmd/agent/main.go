package main

import (
	"flag"
	"fmt"
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
	reportIntervalArg *int
	pollIntervalArg   *int
)

func init() {
	srvAdressArg = flag.String("a", "localhost:8080", "host:port (default localhost:8080)")
	reportIntervalArg = flag.Int("r", 10, "report to server interval in seconds (default 10s)")
	pollIntervalArg = flag.Int("p", 2, "poll interval in seconds (default 2s)")
}

func main() {
	var cfg agent.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	flag.Parse()
	if _, err := os.LookupEnv("ADDRESS"); err {
		cfg.Server = *srvAdressArg
	}
	if _, err := os.LookupEnv("REPORT_INTERVAL"); err {
		cfg.ReportInterval = time.Duration(*reportIntervalArg)
	}
	if _, err := os.LookupEnv("POLL_INTERVAL"); err {
		cfg.PollInterval = time.Duration(*pollIntervalArg)
	}
	fmt.Println(cfg.PollInterval)
	fmt.Println(cfg.ReportInterval)
	fmt.Println(cfg.Server)

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
