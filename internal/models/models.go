package models

import (
	"encoding/json"
	"log"
)

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func (m Metrics) String() string {
	j, err := json.Marshal(m)
	if err != nil {
		log.Println("Stringer error")
		return ""
	}
	ret := string(j)
	return ret[1 : len(ret)-1]
}

func (m Metrics) StringData() string {
	m.Hash = ""
	j, err := json.Marshal(m)
	if err != nil {
		log.Println("Stringer error")
		return ""
	}
	ret := string(j)
	return ret[1 : len(ret)-1]
}
