package storage

import (
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

type Repository struct {
	//	db     map[string]string
	JSONDB []models.Metrics
}

func (r *Repository) InsertMetric(m models.Metrics) error {
	r.AppendMetric(m)
	return nil
}

func (r *Repository) AppendMetric(m models.Metrics) {
	for i := range r.JSONDB {
		if r.JSONDB[i].ID == m.ID {
			newDelta := *(r.JSONDB[i].Delta) + *(m.Delta)
			r.JSONDB[i].Delta = &newDelta
			r.JSONDB[i].Value = m.Value
			return
		}
	}
	//	log.Println(m)
	r.JSONDB = append(r.JSONDB, m)
}

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

func (r *Repository) Restore(file string) {
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

func (r *Repository) GetMetric(data models.Metrics) (models.Metrics, error) {
	for i := range r.JSONDB {
		//log.Printf("db: %s , data:%s", r.JSONDB[i].ID, data.ID)
		if r.JSONDB[i].ID == data.ID {
			data.Value = r.JSONDB[i].Value
			data.Delta = r.JSONDB[i].Delta
			return data, nil
		}
	}
	//log.Println("here")
	var intVal = new(int64)
	floatVal := 0.0
	data.Delta = intVal
	data.Value = &floatVal
	return data, errors.New("no data")

}

func (r *Repository) InsertData(typeVar string, name string, value string) int {
	var model models.Metrics
	model.ID = name
	model.MType = typeVar
	//	log.Println(value)
	if typeVar == "gauge" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseFloat(value, 64)
			model.Value = &tmp
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	if typeVar == "counter" {
		if utils.CheckIfStringIsNumber(value) {
			tmp, _ := strconv.ParseInt(value, 10, 64)
			model.Delta = &tmp
			//	log.Println(*model.Delta)
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	r.InsertMetric(model)
	return http.StatusOK
}

func NewRepo() Repository {
	return Repository{
		JSONDB: []models.Metrics{},
		//	db: make(map[string]string),
	}
}

/*
func (r Repository) GetAll() map[string]string {
	return r.db
}
func (r Repository) GetByName(name string) (string, bool) {
	if value, ok := r.db[name]; ok {
		return value, true
	}
	return "", false
}

func (r Repository) insertGouge(name, value string) error {
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		r.db[name] = value
		return nil
	}
	return errors.New("not float")
}

func (r Repository) insertCount(name, value string) error {
	if oldVal, ok := r.db[name]; ok {
		oldInt, _ := strconv.ParseInt(oldVal, 10, 64)
		newInt, _ := strconv.ParseInt(value, 10, 64)
		r.db[name] = fmt.Sprint(newInt + oldInt)
		return nil
	} else if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		r.db[name] = value
		return nil
	}
	return errors.New("not int")
} */
