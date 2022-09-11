//package models contains data models which is used in app.
package models

import (
	"context"
	"fmt"
)

//Metrics describe metric structure.
type Metrics struct {
	//ID - name of metric from runtime.
	ID string `json:"id"` // имя метрики
	//MType - type of metric (gauge/counter).
	MType string `json:"type"` // параметр, принимающий значение gauge или counter
	//Delta - pointer to counter value (int64).
	Delta *int64 `json:"delta,omitempty"` // значение метрики в случае передачи counter
	//Value - pointer to gauge value (float64).
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	//Hash - MAC.
	Hash string `json:"hash,omitempty"` // значение хеш-функции
}

//MetricsDB - []Metrics, array of metrics.
type MetricsDB []Metrics

//StringData return string "name:type:value" of metric.
func (m *Metrics) StringData() string {
	return m.formatString()
}

func (m *Metrics) formatString() string {
	//log.Println(*m)
	switch m.MType {
	case "gauge":
		//log.Printf("gauge %f", *m.Value)
		return fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	case "counter":
		return fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	}
	return ""
}

//NewMetrics - models.Metrics constructor.
func NewMetric(varID string, varType string, varDelta *int64, varValue *float64, varHash string) Metrics {
	return Metrics{
		ID:    varID,
		MType: varType,
		Delta: varDelta,
		Value: varValue,
		Hash:  varHash,
	}

}

//Storager - Interface which is used app to save the data.
type Storager interface {
	//InsertMetric - save models.Metrics.
	InsertMetric(ctx context.Context, m Metrics) error
	//GetMetric - get model.Metrics from storage.
	GetMetric(data Metrics) (Metrics, error)
	//InsertData - save metric raw data.
	InsertData(ctx context.Context, typeVar string, name string, value string, hash string) int
	//GetAll - get all model.Metrics data from storage.
	GetAll(ctx context.Context) []Metrics
	//SaveData - save data from storage to file.
	SaveData(file string)
	//Restore - restore data from file to storage.
	Restore(file string)
	//PingDB - get state of current storage.
	PingDB() bool
	//BatchInsert - Insert all collected metrics in one batch.
	BatchInsert(ctx context.Context, dataModels []Metrics) error
	//GetCurrentCommit - Get current commit from storage.
	GetCurrentCommit() float64
}
