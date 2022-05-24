package storage

import (
	"fmt"
	"strconv"
)

type Storager interface {
	InsertGouge(name string, value string) error
}

type Repository struct {
	db map[string]string
}

func NewRepo() Repository {
	return Repository{
		db: make(map[string]string),
	}
}

func (r Repository) InsertGouge(name, value string) error {
	r.db[name] = value
	return nil
}

func (r Repository) InsertCount(name, value string) error {
	if old_val, ok := r.db[name]; ok {
		old_int, _ := strconv.ParseInt(old_val, 10, 64)
		new_int, _ := strconv.ParseInt(value, 10, 64)
		r.db[name] = fmt.Sprint(new_int + old_int)
	} else {
		r.db[name] = value
	}

	return nil
}
