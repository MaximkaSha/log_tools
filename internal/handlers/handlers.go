package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	handlers *http.ServeMux
	Repo     storage.Repository
	SyncFile string
}

func NewHandlers(repo storage.Repository) Handlers {
	handl := http.NewServeMux()
	return Handlers{
		handlers: handl,
		Repo:     repo,
		SyncFile: "",
	}
}

func (obj *Handlers) HandleUpdate(w http.ResponseWriter, r *http.Request) { //should be renamed to HandlePostUpdate

	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	valueVal := chi.URLParam(r, "value")
	w.Header().Set("Content-Type", "text/plain")
	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	result := obj.Repo.InsertData(typeVal, nameVal, valueVal)
	if result != 200 {
		http.Error(w, "Bad value found!", result)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
curl --header "Content-Type: application/json" --request POST --data "{\"id\":\"PollCount\",\"type\":\"gauge\",\"value\":10.0230}" http://localhost:8080/update/
*/

func (obj *Handlers) HandlePostJSONUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data = new(models.Metrics)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		obj.Repo.InsertMetric(*data)
		//obj.Repo.SaveData(obj.SyncFile)
		w.WriteHeader(http.StatusOK)
		jData, _ := json.Marshal(data)
		w.Write(jData)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

func (obj *Handlers) HandlePostJSONValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data = new(models.Metrics)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, "Data error!", http.StatusBadRequest)
			return
		}
		for i := range obj.Repo.JSONDB {
			if obj.Repo.JSONDB[i].ID == data.ID {
				data.Value = obj.Repo.JSONDB[i].Value
				data.Delta = obj.Repo.JSONDB[i].Delta
				jData, _ := json.Marshal(data)
				w.WriteHeader(http.StatusOK)
				w.Write(jData)
				return
			}
		}
		if val, ok := obj.Repo.GetByName(data.ID); ok {
			if data.MType != "gauge" {
				intVal, _ := strconv.ParseInt(val, 10, 64)
				data.Delta = &intVal
			} else if data.MType != "counter" {
				floatVal, _ := strconv.ParseFloat(val, 64)
				data.Value = &floatVal
			} else {
				http.Error(w, "Type not found!", http.StatusNotImplemented)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			jData, _ := json.Marshal(data)
			w.WriteHeader(http.StatusOK)
			w.Write(jData)
		} else {
			var intVal = new(int64)
			floatVal := 0.0
			data.Delta = intVal
			data.Value = &floatVal
			w.Header().Set("Content-Type", "application/json")
			jData, _ := json.Marshal(data)
			w.WriteHeader(http.StatusNotFound)
			w.Write(jData)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, "Not json", http.StatusBadRequest)
		return

	}

}

func (obj *Handlers) HandleGetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	repo := obj.Repo.JSONDB
	allData, _ := json.MarshalIndent(repo, "", "    ")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(allData))
}

func (obj *Handlers) HandleGetUpdate(w http.ResponseWriter, r *http.Request) {
	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")

	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	if valueVar, ok := obj.Repo.GetByName(nameVal); !ok {
		http.Error(w, "Name not found!", http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(valueVar))
	}

}
