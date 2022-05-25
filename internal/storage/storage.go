package storage

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/utils"
)

type Storager interface {
	InsertData(typeVar string, name string, value string) int
}

type Repository struct {
	db map[string]string
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
	//fmt.Println(name)
	//fmt.Println(r.db[name])
	if value, ok := r.db[name]; ok {
		return value, true
	}
	return "", false
}

func (r Repository) insertGouge(name, value string) error {
	r.db[name] = value
	return nil
}

func (r Repository) insertCount(name, value string) error {
	if oldVal, ok := r.db[name]; ok {
		oldInt, _ := strconv.ParseInt(oldVal, 10, 64)
		newInt, _ := strconv.ParseInt(value, 10, 64)
		r.db[name] = fmt.Sprint(newInt + oldInt)
	} else {
		r.db[name] = value
	}

	return nil
}
