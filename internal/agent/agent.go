package agent

import (
	"bytes"
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

	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/models"
)

type Config struct {
	Server         string        `env:"ADDRESS" envDefault:"localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL,required" envDefault:"2s"`
	KeyFile        string        `env:"KEY" envDefault:"key.txt"`
}

type Agent struct {
	logDB   []models.Metrics
	counter int64
	cfg     Config
}

func NewAgent(cfg Config) Agent {
	return Agent{
		logDB:   []models.Metrics{},
		counter: 0,
		cfg:     cfg,
	}
}

func (a *Agent) AppendMetric(m models.Metrics) {
	for i := range a.logDB {
		if a.logDB[i].ID == m.ID {
			a.logDB[i].Delta = m.Delta
			a.logDB[i].Value = m.Value
			return
		}
	}
	a.logDB = append(a.logDB, m)
}

func (a *Agent) StartService() {
	var pollInterval = a.cfg.PollInterval
	var reportInterval = a.cfg.ReportInterval
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
			a.CollectLogs()
		case <-tickerSend.C:
			go a.AgentSendWorker()
		case <-sigc:
			log.Println("Got quit signal.")
			return
		}
	}
}

//Надо передавать контекст, что убивать рутину, если началась новая или убить по требыванию
func (a Agent) AgentSendWorker() {
	a.SendLogsbyPost("http://" + a.cfg.Server + "/update/")
	a.SendLogsbyJSON("http://" + a.cfg.Server + "/update/")
	a.SendLogsbyJSONBatch("http://" + a.cfg.Server + "/updates/")
}

func (a Agent) SendLogsbyJSONBatch(url string) error {
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	var allData = []models.Metrics{}
	for i := range a.logDB {
		var data = models.Metrics{}
		data = a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}
		}
		allData = append(allData, data)
	}
	jData, _ := json.Marshal(allData)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jData))
	defer resp.Body.Close()
	if err != nil {
		log.Println("error sending logs")
	}
	log.Println("Sended logs by POST JSON Batch")
	return nil
}

func (a Agent) SendLogsbyJSON(url string) error {
	hasher := crypto.NewCryptoService()
	hasher.InitCryptoService(a.cfg.KeyFile)
	for i := range a.logDB {
		var data = models.Metrics{}
		data = a.logDB[i]
		if hasher.IsServiceEnable() {
			_, err := hasher.Hash(&data)
			if err != nil {
				log.Println("Hasher error!")
				continue
			}
		}
		//log.Println(data)
		jData, _ := json.Marshal(data)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jData))
		if err == nil {
			resp.Body.Close()
		}
	}
	log.Println("Sended logs by POST JSON")
	//log.Println(a.logDB)
	return nil
}

func (a Agent) getPostStrByIndex(i int, url string) string {
	if a.logDB[i].MType == "counter" {
		return fmt.Sprintf(url+"%s/%s/%d", a.logDB[i].MType, a.logDB[i].ID, *a.logDB[i].Delta)
	} else if a.logDB[i].MType == "gauge" {
		return fmt.Sprintf(url+"%s/%s/%f", a.logDB[i].MType, a.logDB[i].ID, *a.logDB[i].Value)
	}
	return "type unknown"
}

func (a *Agent) SendLogsbyPost(sData string) error {
	for i := range a.logDB {
		//TODO: make config struct part of agent class
		if r, err := http.Post(a.getPostStrByIndex(i, sData), "text/plain", nil); err == nil {
			r.Body.Close()
		}
	}
	log.Println("Sended logs by POST param")
	//log.Println(a.logDB)
	return nil
}

func (a *Agent) CollectLogs() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	var tmpAlloc = float64(rtm.Alloc)
	a.AppendMetric(models.Metrics{ID: "Alloc", MType: "gauge", Delta: nil, Value: &tmpAlloc})
	var tmpBuckHashSys = float64(rtm.BuckHashSys)
	a.AppendMetric(models.Metrics{ID: "BuckHashSys", MType: "gauge", Delta: nil, Value: &tmpBuckHashSys})
	var tmpFrees = float64(rtm.Frees)
	a.AppendMetric(models.Metrics{ID: "Frees", MType: "gauge", Delta: nil, Value: &tmpFrees})
	var tmpGCCPUFraction = float64(rtm.GCCPUFraction)
	a.AppendMetric(models.Metrics{ID: "GCCPUFraction", MType: "gauge", Delta: nil, Value: &tmpGCCPUFraction})
	var tmpGCSys = float64(rtm.GCSys)
	a.AppendMetric(models.Metrics{ID: "GCSys", MType: "gauge", Delta: nil, Value: &tmpGCSys})
	var tmpHeapAlloc = float64(rtm.HeapAlloc)
	a.AppendMetric(models.Metrics{ID: "HeapAlloc", MType: "gauge", Delta: nil, Value: &tmpHeapAlloc})
	var tmpHeapIdle = float64(rtm.HeapIdle)
	a.AppendMetric(models.Metrics{ID: "HeapIdle", MType: "gauge", Delta: nil, Value: &tmpHeapIdle})
	var tmpHeapInuse = float64(rtm.HeapInuse)
	a.AppendMetric(models.Metrics{ID: "HeapInuse", MType: "gauge", Delta: nil, Value: &tmpHeapInuse})
	var tmpHeapObject = float64(rtm.HeapObjects)
	a.AppendMetric(models.Metrics{ID: "HeapObjects", MType: "gauge", Delta: nil, Value: &tmpHeapObject})
	var tmpHeapReleased = float64(rtm.HeapReleased)
	a.AppendMetric(models.Metrics{ID: "HeapReleased", MType: "gauge", Delta: nil, Value: &tmpHeapReleased})
	var tmpHeapSys = float64(rtm.HeapSys)
	a.AppendMetric(models.Metrics{ID: "HeapSys", MType: "gauge", Delta: nil, Value: &tmpHeapSys})
	var tmpLastGC = float64(rtm.LastGC)
	a.AppendMetric(models.Metrics{ID: "LastGC", MType: "gauge", Delta: nil, Value: &tmpLastGC})
	var tmpLookups = float64(rtm.Lookups)
	a.AppendMetric(models.Metrics{ID: "Lookups", MType: "gauge", Delta: nil, Value: &tmpLookups})
	var tmpMCacheInuse = float64(rtm.MCacheInuse)
	a.AppendMetric(models.Metrics{ID: "MCacheInuse", MType: "gauge", Delta: nil, Value: &tmpMCacheInuse})
	var tmpMCacheSys = float64(rtm.MCacheSys)
	a.AppendMetric(models.Metrics{ID: "MCacheSys", MType: "gauge", Delta: nil, Value: &tmpMCacheSys})
	var tmpMSpanInuse = float64(rtm.MSpanInuse)
	a.AppendMetric(models.Metrics{ID: "MSpanInuse", MType: "gauge", Delta: nil, Value: &tmpMSpanInuse})
	var tmpMSpanSys = float64(rtm.MSpanSys)
	a.AppendMetric(models.Metrics{ID: "MSpanSys", MType: "gauge", Delta: nil, Value: &tmpMSpanSys})
	var tmpNextGC = float64(rtm.NextGC)
	a.AppendMetric(models.Metrics{ID: "NextGC", MType: "gauge", Delta: nil, Value: &tmpNextGC})
	var tmpOtherSys = float64(rtm.OtherSys)
	a.AppendMetric(models.Metrics{ID: "OtherSys", MType: "gauge", Delta: nil, Value: &tmpOtherSys})
	var tmpPauseTotalNs = float64(rtm.PauseTotalNs)
	a.AppendMetric(models.Metrics{ID: "PauseTotalNs", MType: "gauge", Delta: nil, Value: &tmpPauseTotalNs})
	var tmpRandomValue = float64(rand.Int63())
	a.AppendMetric(models.Metrics{ID: "RandomValue", MType: "gauge", Delta: nil, Value: &tmpRandomValue})
	a.counter = a.counter + 1
	tmpI := a.counter
	a.AppendMetric(models.Metrics{ID: "PollCount", MType: "counter", Delta: &tmpI, Value: nil})
	var tmpNumForcedGC = float64(rtm.NumForcedGC)
	a.AppendMetric(models.Metrics{ID: "NumForcedGC", MType: "gauge", Delta: nil, Value: &tmpNumForcedGC})
	var tmpNumGC = float64(rtm.NumGC)
	a.AppendMetric(models.Metrics{ID: "NumGC", MType: "gauge", Delta: nil, Value: &tmpNumGC})
	var tmpStackInuse = float64(rtm.StackInuse)
	a.AppendMetric(models.Metrics{ID: "StackInuse", MType: "gauge", Delta: nil, Value: &tmpStackInuse})
	var tmpStackSys = float64(rtm.StackSys)
	a.AppendMetric(models.Metrics{ID: "StackSys", MType: "gauge", Delta: nil, Value: &tmpStackSys})
	var tmpSys = float64(rtm.Sys)
	a.AppendMetric(models.Metrics{ID: "Sys", MType: "gauge", Delta: nil, Value: &tmpSys})
	var tmpTotalAlloc = float64(rtm.TotalAlloc)
	a.AppendMetric(models.Metrics{ID: "TotalAlloc", MType: "gauge", Delta: nil, Value: &tmpTotalAlloc})
	var tmpMallocs = float64(rtm.Mallocs)
	a.AppendMetric(models.Metrics{ID: "Mallocs", MType: "gauge", Delta: nil, Value: &tmpMallocs})

	log.Println("Collected logs")
	//	log.Println(a.logDB)
}

/*
Mallocs
NumForcedGC
NumGC
StackInuse
StackSys
Sys
TotalAlloc
*/
