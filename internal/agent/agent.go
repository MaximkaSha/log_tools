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

func (a Agent) SendLogsbyJson(url string) error {
	for i := range a.logDB {
		var data = models.Metrics{}
		data = a.logDB[i]
		jData, _ := json.Marshal(data)
		log.Println(url)
		http.Post(url, "application/json", bytes.NewBuffer(jData))
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
		log.Println(a.getPostStrByIndex(i, sData))
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
	var tmp = float64(rtm.Alloc)
	a.AppendMetric(models.Metrics{"Alloc", "gauge", nil, &tmp})
	tmp = float64(rtm.BuckHashSys)
	a.AppendMetric(models.Metrics{"BuckHashSys", "gauge", nil, &tmp})
	tmp = float64(rtm.Frees)
	a.AppendMetric(models.Metrics{"Frees", "gauge", nil, &tmp})
	tmp = float64(rtm.GCCPUFraction)
	a.AppendMetric(models.Metrics{"GCCPUFraction", "gauge", nil, &tmp})
	tmp = float64(rtm.GCSys)
	a.AppendMetric(models.Metrics{"GCSys", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapAlloc)
	a.AppendMetric(models.Metrics{"HeapAlloc", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapIdle)
	a.AppendMetric(models.Metrics{"HeapIdle", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapInuse)
	a.AppendMetric(models.Metrics{"HeapInuse", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapObjects)
	a.AppendMetric(models.Metrics{"HeapObjects", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapReleased)
	a.AppendMetric(models.Metrics{"HeapReleased", "gauge", nil, &tmp})
	tmp = float64(rtm.HeapSys)
	a.AppendMetric(models.Metrics{"HeapSys", "gauge", nil, &tmp})
	tmp = float64(rtm.LastGC)
	a.AppendMetric(models.Metrics{"LastGC", "gauge", nil, &tmp})
	tmp = float64(rtm.Lookups)
	a.AppendMetric(models.Metrics{"Lookups", "gauge", nil, &tmp})
	tmp = float64(rtm.MCacheInuse)
	a.AppendMetric(models.Metrics{"MCacheInuse", "gauge", nil, &tmp})
	tmp = float64(rtm.MCacheSys)
	a.AppendMetric(models.Metrics{"MCacheSys", "gauge", nil, &tmp})
	tmp = float64(rtm.MSpanInuse)
	a.AppendMetric(models.Metrics{"MSpanInuse", "gauge", nil, &tmp})
	tmp = float64(rtm.MSpanSys)
	a.AppendMetric(models.Metrics{"MSpanSys", "gauge", nil, &tmp})
	tmp = float64(rtm.NextGC)
	a.AppendMetric(models.Metrics{"NextGC", "gauge", nil, &tmp})
	tmp = float64(rtm.OtherSys)
	a.AppendMetric(models.Metrics{"OtherSys", "gauge", nil, &tmp})
	tmp = float64(rtm.PauseTotalNs)
	a.AppendMetric(models.Metrics{"PauseTotalNs", "gauge", nil, &tmp})
	tmp = float64(rand.Int63())
	a.AppendMetric(models.Metrics{"RandomValue", "gauge", nil, &tmp})
	a.counter = a.counter + 1
	tmpI := a.counter
	a.AppendMetric(models.Metrics{"PollCounter", "counter", &tmpI, nil})
	log.Println("Collected logs")
	//	log.Println(a.logDB)
}
