package models

import (
	"fmt"
	"log"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

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
	log.Println(*m)
	switch m.MType {
	case "gauge":
		//log.Printf("gauge %f", *m.Value)
		return fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	case "counter":
		return fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	}
	return ""
}

type Storager interface {
	InsertMetric(m Metrics) error
	GetMetric(data Metrics) (Metrics, error)
	InsertData(typeVar string, name string, value string, hash string) int
	GetAll() []Metrics
	SaveData(file string)
	Restore(file string)
	PingDB() bool
}
