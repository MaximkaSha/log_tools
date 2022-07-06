package models

import (
	"context"
	"fmt"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

type MetricsDB []Metrics

func (m *Metrics) StringData() string {
	/*	m.Hash = ""
		j, err := json.Marshal(m)
		if err != nil {
			log.Println("Stringer error")
			return ""
		}
		ret := string(j) */
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
func NewMetric(varID string, varType string, varDelta *int64, varValue *float64, varHash string) Metrics {
	return Metrics{
		ID:    varID,
		MType: varType,
		Delta: varDelta,
		Value: varValue,
		Hash:  varHash,
	}

}

type Storager interface {
	InsertMetric(ctx context.Context, m Metrics) error
	GetMetric(data Metrics) (Metrics, error)
	InsertData(ctx context.Context, typeVar string, name string, value string, hash string) int
	GetAll(ctx context.Context) []Metrics
	SaveData(file string)
	Restore(file string)
	PingDB() bool
	BatchInsert(ctx context.Context, dataModels []Metrics) error
	GetCurrentCommit() float64
}
