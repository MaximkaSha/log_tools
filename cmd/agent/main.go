package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/MaximkaSha/log_tools/internal/agent"
)

type gauge float64
type counter int64

type logData struct {
	Alloc         gauge
	BuckHashSys   gauge
	Frees         gauge
	GCCPUFraction gauge
	GCSys         gauge
	HeapAlloc     gauge
	HeapIdle      gauge
	HeapInuse     gauge
	HeapObjects   gauge
	HeapReleased  gauge
	HeapSys       gauge
	LastGC        gauge
	Lookups       gauge
	MCacheInuse   gauge
	MCacheSys     gauge
	MSpanInuse    gauge
	MSpanSys      gauge
	Mallocs       gauge
	NextGC        gauge
	NumForcedGC   gauge
	NumGC         gauge
	OtherSys      gauge
	PauseTotalNs  gauge
	StackInuse    gauge
	StackSys      gauge
	Sys           gauge
	TotalAlloc    gauge
	PollCount     counter
	RandomValue   gauge
}

func collectLogs(ld *logData, rtm runtime.MemStats) int64 {
	runtime.ReadMemStats(&rtm)
	ld.Alloc = gauge(rtm.Alloc)
	ld.BuckHashSys = gauge(rtm.BuckHashSys)
	ld.Frees = gauge(rtm.Frees)
	ld.GCCPUFraction = gauge(rtm.GCCPUFraction)
	ld.GCSys = gauge(rtm.GCSys)
	ld.HeapAlloc = gauge(rtm.HeapAlloc)
	ld.HeapIdle = gauge(rtm.HeapIdle)
	ld.HeapInuse = gauge(rtm.HeapInuse)
	ld.HeapObjects = gauge(rtm.HeapObjects)
	ld.HeapReleased = gauge(rtm.HeapReleased)
	ld.HeapSys = gauge(rtm.HeapSys)
	ld.LastGC = gauge(rtm.LastGC)
	ld.Lookups = gauge(rtm.Lookups)
	ld.MCacheInuse = gauge(rtm.MCacheInuse)
	ld.MCacheSys = gauge(rtm.MCacheSys)
	ld.MSpanInuse = gauge(rtm.MSpanInuse)
	ld.MSpanSys = gauge(rtm.MSpanSys)
	ld.Mallocs = gauge(rtm.Mallocs)
	ld.NextGC = gauge(rtm.NextGC)
	ld.OtherSys = gauge(rtm.OtherSys)
	ld.PauseTotalNs = gauge(rtm.PauseTotalNs)
	ld.PollCount = counter(ld.PollCount + 1)
	ld.RandomValue = gauge(rand.Int63())

	log.Printf("data #%d collected with rnd %x", ld.PollCount, ld.RandomValue)
	return int64(ld.RandomValue)

}

func sendLogs(ld logData) {
	jjson, _ := json.Marshal(ld)
	var x map[string]interface{}
	_ = json.Unmarshal(jjson, &x)
	var postStr string
	for k, v := range x {
		postStr = ""
		if k == "PollCount" {
			postStr = fmt.Sprintf("http://127.0.0.1:8080/update/counter/%s/%.f", k, v)
			fmt.Println(postStr)
		} else {
			postStr = fmt.Sprintf("http://127.0.0.1:8080/update/gauge/%s/%f", k, v)
		}
		if r, err := http.Post(postStr, "text/plain", nil); err == nil {
			r.Body.Close()
		}
		//	log.Printf("Transfer data %s", postStr)
	}
	log.Printf("Sended data #%d with rnd %x", ld.PollCount, ld.RandomValue)
}

func main() {
	agentService := agent.NewAgent()
	var pollInterval = 2 * time.Second
	var reportInterval = 10 * time.Second
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
			agentService.SendLogsbyPost("http://localhost:8080/update/")
			agentService.SendLogsbyJson("http://localhost:8080/update/")
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}

}
