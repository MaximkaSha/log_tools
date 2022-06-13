package storage

import (
	"encoding/json"
	"errors"
	"fmt"

	//"json/encoding"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/utils"
)

type Repository struct {
	db     map[string]string
	JSONDB []models.Metrics
}

func (r *Repository) InsertMetric(m models.Metrics) error {
	if m.MType == "counter" {
		if oldVal, ok := r.db[m.ID]; ok {
			oldInt, _ := strconv.ParseInt(oldVal, 10, 64)
			newInt := *m.Delta
			r.db[m.ID] = fmt.Sprint(newInt + oldInt)
			tmpVar := newInt + oldInt
			m.Delta = &tmpVar
			log.Println(m)
		} else {
			r.db[m.ID] = fmt.Sprint(m.Delta)
		}
	}
	if m.MType == "gauge" {
		r.db[m.ID] = fmt.Sprint(*m.Value)
	}
	r.AppendMetric(m)
	return nil
}

func (r *Repository) AppendMetric(m models.Metrics) {

	for i := range r.JSONDB {
		if r.JSONDB[i].ID == m.ID && m.MType == "counter" {
			newDelta := *(r.JSONDB[i].Delta) + *(m.Delta)
			r.JSONDB[i].Delta = &newDelta
			r.JSONDB[i].Value = m.Value
			return
		} else if r.JSONDB[i].ID == m.ID && m.MType == "gauge" {
			r.JSONDB[i].Delta = m.Delta
			r.JSONDB[i].Value = m.Value
			return
		}
	}
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
	var data []models.Metrics
	var jData, err = ioutil.ReadFile(file)
	if err != nil {
		log.Panic(err)
	}
	err = json.Unmarshal(jData, &data)
	if err != nil {
		log.Panic(err)
	}
	r.JSONDB = data
	log.Print("Data restored from file")

}

func (r Repository) InsertData(typeVar string, name string, value string) int {
	if typeVar == "gauge" {
		if utils.CheckIfStringIsNumber(value) {
			r.insertGouge(name, value)
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	if typeVar == "counter" {
		if utils.CheckIfStringIsNumber(value) {
			r.insertCount(name, value)
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func NewRepo() Repository {
	return Repository{
		db: make(map[string]string),
	}
}

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
}
