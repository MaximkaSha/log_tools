package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	handlers *http.ServeMux
	repo     storage.Repository
}

func NewHandlers(repo storage.Repository) Handlers {
	handl := http.NewServeMux()
	return Handlers{
		handlers: handl,
		repo:     repo,
	}
}

func (obj Handlers) HandleUpdate(w http.ResponseWriter, r *http.Request) { //should be renamed to HandlePostUpdate

	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	valueVal := chi.URLParam(r, "value")

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed) //legacy, before chi refactoring
		return                                                                        // should be deleted
	}

	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	result := obj.repo.InsertData(typeVal, nameVal, valueVal)
	if result != 200 {
		http.Error(w, "Bad value found!", result)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
curl --header "Content-Type: application/json" --request POST --data "{\"id\":\"PollCount\",\"type\":\"gauge\",\"value\":10.0230}" http://localhost:8080/update/
*/

func (obj Handlers) HandlePostJSONUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data = new(models.Metrics)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		obj.repo.InsertMetric(*data)
		w.WriteHeader(http.StatusOK)
		//	obj.repo.GetByName(data)
		jData, _ := json.Marshal(data)
		w.Write(jData)
	} else {
		//fmt.Println(r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNotFound)
	}

}

func (obj Handlers) HandlePostJSONValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data = new(models.Metrics)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "Data error!", http.StatusBadRequest)
		}
		if val, ok := obj.repo.GetByName(data.ID); ok {
			if data.MType != "gauge" {
				*data.Delta, _ = strconv.ParseInt(val, 10, 64)
			} else if data.MType != "counter" {
				*data.Value, _ = strconv.ParseFloat(val, 64)
			} else {
				http.Error(w, "Type not found!", http.StatusNotImplemented)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			jData, _ := json.Marshal(data)
			w.WriteHeader(http.StatusOK)
			w.Write(jData)
		} else {
			*data.Delta = 0
			*data.Value = 0
			w.Header().Set("Content-Type", "application/json")
			jData, _ := json.Marshal(data)
			w.WriteHeader(http.StatusNotFound)
			w.Write(jData)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Not json", http.StatusBadRequest)

	}

}

func (obj Handlers) HandleGetHome(w http.ResponseWriter, r *http.Request) {
	repo := obj.repo.GetAll()
	var allData = ""
	for key, value := range repo {
		s := fmt.Sprintf("%s = %s\n", key, value)
		allData = allData + s
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(allData))
}

func (obj Handlers) HandleGetUpdate(w http.ResponseWriter, r *http.Request) {
	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")

	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	if valueVar, ok := obj.repo.GetByName(nameVal); !ok {
		http.Error(w, "Name not found!", http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valueVar))
	}

}
