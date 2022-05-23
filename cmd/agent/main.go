package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"
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

func collectLogs(ld *logData, rtm runtime.MemStats) {
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

}

func sendLogs(ld logData) {
	v := reflect.ValueOf(ld)
	typeOfS := v.Type()
	var postStr string
	for i := 0; i < v.NumField(); i++ {
		postStr = ""
		typeVar := typeOfS.Field(i).Type.String()
		nameVar := typeOfS.Field(i).Name
		valueVar := v.Field(i).Interface()
		//	fmt.Println(valueVar)
		//	fmt.Println(nameVar)
		//	fmt.Println(typeVar[5:])
		if typeVar[5:] == "gauge" {
			postStr = fmt.Sprintf("http://127.0.0.1:8080/update/%s/%s/%f", typeVar[5:], nameVar, valueVar)
		} else {
			postStr = fmt.Sprintf("http://127.0.0.1:8080/update/%s/%s/%d", typeVar[5:], nameVar, valueVar)
		}
		//fmt.Println(postStr)
		http.Post(postStr, "text/plain", nil)
		log.Printf("Transfer data %s", postStr)
	}
	log.Printf("Sended data #%d with rnd %x", ld.PollCount, ld.RandomValue)
}

func main() {
	var pollInterval = 2 * time.Second
	var reportInterval = 10 * time.Second
	var logData = new(logData)
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	var rtm runtime.MemStats
	log.Println("Logger start...")
	tickerCollect := time.NewTicker(pollInterval)
	tickerSend := time.NewTicker(reportInterval)
	defer tickerCollect.Stop()
	defer tickerSend.Stop()
	for {
		select {
		case <-tickerCollect.C:
			collectLogs(logData, rtm)
		case <-tickerSend.C:
			sendLogs(*logData)
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}

}
