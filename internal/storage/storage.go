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

func (repo Repository) InsertData(typeVar string, name string, value string) int {
	if typeVar == "gauge" {
		if utils.CheckIfStringIsNumber(value) {
			repo.insertGouge(name, value)
		} else {
			//http.Error(w, "Bad value found!", http.StatusBadRequest)
			return http.StatusBadRequest
		}
	}
	if typeVar == "counter" {
		if utils.CheckIfStringIsNumber(value) {
			repo.insertCount(name, value)
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

func (r Repository) insertGouge(name, value string) error {
	r.db[name] = value
	return nil
}

func (r Repository) insertCount(name, value string) error {
	if old_val, ok := r.db[name]; ok {
		old_int, _ := strconv.ParseInt(old_val, 10, 64)
		new_int, _ := strconv.ParseInt(value, 10, 64)
		r.db[name] = fmt.Sprint(new_int + old_int)
	} else {
		r.db[name] = value
	}

	return nil
}
