package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/agent"
	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg agent.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
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
			//agentService.SendLogsbyPost("http://localhost:8080/update/")
			agentService.SendLogsbyJSON("http://" + cfg.Server + "/update/")
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}

}
