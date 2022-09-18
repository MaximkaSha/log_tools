// package storage provide in-memory storage for app.
package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/utils"
)

// Repository - in memory storage.
type Repository struct {
	// JSONDB - array of models.Metrics
	JSONDB []models.Metrics
}

// InsertMetrics - add models.Metrics to storage.
func (r *Repository) InsertMetric(ctx context.Context, m models.Metrics) error {
	r.AppendMetric(m)
	return nil
}

// AppendMetric - add models.Metrics to storage.
//
// DEPRICATED: use InsertMetric.
func (r *Repository) AppendMetric(m models.Metrics) {
	for i := range r.JSONDB {
		if r.JSONDB[i].ID == m.ID {
			if m.Delta != nil {
				newDelta := *(r.JSONDB[i].Delta) + *(m.Delta)
				r.JSONDB[i].Delta = &newDelta
			}
			r.JSONDB[i].Value = m.Value
			r.JSONDB[i].Hash = m.Hash
			return
		}
	}
	r.JSONDB = append(r.JSONDB, m)
}

// SaveData - save data from in-memory storage to file.
func (r *Repository) SaveData(file string) {
	if file == "" {
		return
	}
	jData, err := json.Marshal(r.JSONDB)
	if err != nil {
		log.Panic(err)
	}
	_ = ioutil.WriteFile(file, jData, 0644)
}

// Restore - restore data from file to in-memory storage.
func (r *Repository) Restore(file string) {
	log.Println(file)
	if _, err := os.Stat(file); err != nil {
		log.Println("Restore file not found")
		return
	}
	var data []models.Metrics
	var jData, err = ioutil.ReadFile(file)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(jData, &data)
	if err != nil {
		log.Println("Data file corrupted")
	} else {
		r.JSONDB = data
		log.Print("Data restored from file")
	}

}

// GetMetric - get models.Metrics from storage.
func (r *Repository) GetMetric(data models.Metrics) (models.Metrics, error) {
	for i := range r.JSONDB {
		if r.JSONDB[i].ID == data.ID {
			data.Value = r.JSONDB[i].Value
			data.Delta = r.JSONDB[i].Delta
			return data, nil
		}
	}
	var intVal = new(int64)
	floatVal := 0.0
	data.Delta = intVal
	data.Value = &floatVal
	return data, errors.New("no data")

}

// InsertData - save raw data (models.Metrics data) to storage.
func (r *Repository) InsertData(ctx context.Context, typeVar string, name string, value string, hash string) int {
	var model models.Metrics
	model.ID = name
	model.MType = typeVar
	// 	log.Println(value)
	if typeVar == "gauge" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseFloat(value, 64)
			model.Value = &tmp
		} else {
			return http.StatusBadRequest
		}
	}
	if typeVar == "counter" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseInt(value, 10, 64)
			model.Delta = &tmp
		} else {
			return http.StatusBadRequest
		}
	}
	model.Hash = hash
	r.InsertMetric(ctx, model)
	return http.StatusOK
}

// GetAll - get all []models.Metrics from storage.
func (r Repository) GetAll(ctx context.Context) []models.Metrics {
	return r.JSONDB
}

// PingDB - get current status of DB.
// Always false (we are not using DB).
func (r Repository) PingDB() bool {
	return false
}

// BatchInsert - not implemented.
func (r Repository) BatchInsert(ctx context.Context, dataModels []models.Metrics) error {
	return errors.New("not implemented for RAM storage")
}

// GetCurrentCommit - return randVal from storage.
func (r Repository) GetCurrentCommit() float64 {
	randVal := models.Metrics{
		ID: "RandomValue",
	}
	randVal, err := r.GetMetric(randVal)
	if err != nil {
		return 0
	}
	return *randVal.Value
}

// NewRepo - Repository constructor.
func NewRepo() Repository {
	return Repository{
		JSONDB: []models.Metrics{},
	}
}
