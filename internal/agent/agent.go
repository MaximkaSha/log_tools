package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"

	"github.com/MaximkaSha/log_tools/internal/models"
)

type Agent struct {
	logDB   []models.Metrics
	counter int64
}

func NewAgent() Agent {
	return Agent{
		logDB:   []models.Metrics{},
		counter: 0,
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

func (a Agent) SendLogsbyJSON(url string) error {
	for i := range a.logDB {
		var data = models.Metrics{}
		data = a.logDB[i]
		jData, _ := json.Marshal(data)
		//	log.Println(url)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jData))
		if err == nil {
			resp.Body.Close()
		}
	}
	log.Println("Sended logs")
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
		//	log.Println(a.getPostStrByIndex(i, sData))
		if r, err := http.Post(a.getPostStrByIndex(i, sData), "text/plain", nil); err == nil {
			r.Body.Close()
		}
	}
	//log.Println("Sended logs")
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
