// Package handlers empelements endpoint api service for agent.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/database"
	"github.com/MaximkaSha/log_tools/internal/models"
	"github.com/go-chi/chi/v5"
)

// Handlers struct.
type Handlers struct {
	Repo          models.Storager
	handlers      *http.ServeMux
	DB            *database.Database
	SyncFile      string
	cryptoService crypto.CryptoService
}

//  NewHandlers constrcutor for Handlers.
func NewHandlers(repo models.Storager, cryptoService crypto.CryptoService) Handlers {
	handl := http.NewServeMux()
	return Handlers{
		handlers:      handl,
		Repo:          repo,
		SyncFile:      "",
		cryptoService: cryptoService,
	}
}

//  HandleUpdate endpoint for raw data.
//  Endpoint get data from URL parametrs type/name/value.
//  Readed data pushed to storage.
//  If type is not gauge or counter, then 501 error.
//  If data is not int64 or float64 then error.
//  If all OK then 200.
func (h *Handlers) HandleUpdate(w http.ResponseWriter, r *http.Request) { // should be renamed to HandlePostUpdate
	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	valueVal := chi.URLParam(r, "value")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	var data models.Metrics
	if h.cryptoService.IsServiceEnable() {
		data.ID = nameVal
		data.MType = typeVal
		switch data.MType {
		case "gauge":
			tmp, _ := strconv.ParseFloat(valueVal, 64)
			data.Value = &tmp
		case "counter":
			tmp, _ := strconv.ParseInt(valueVal, 10, 64)
			data.Delta = &tmp
		}
		h.cryptoService.Hash(&data)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result := h.Repo.InsertData(ctx, typeVal, nameVal, valueVal, data.Hash)
	if result != 200 {
		http.Error(w, "Bad value found!", result)
		return
	}
	w.WriteHeader(http.StatusOK)
}

/*
curl --header "Content-Type: application/json" --request POST --data "{\"id\":\"PollCount\",\"type\":\"gauge\",\"value\":10.0230}" http:// localhost:8080/update/
*/
//  HandlePostJSONUpdate endpoint for JSON data.
//  Endpoint get data from POSTed JSON models.Metrics.
//  Readed data pushed to storage.
//  If type is not gauge or counter, then 501 error.
//  If data is not int64 or float64 then error.
//  If all OK then 200.
func (h *Handlers) HandlePostJSONUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data = new(models.Metrics)
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if h.cryptoService.IsEnable {
			if !h.cryptoService.CheckHash(*data) {
				log.Println("Sing check fail!")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		h.Repo.InsertMetric(ctx, *data)
		// h.Repo.SaveData(h.SyncFile)
		w.WriteHeader(http.StatusOK)
		jData, _ := json.Marshal(data)
		w.Write(jData)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

//  HandlePostJSONValue endpoint for getting data from storage.
//  Endpoint reads models.Metrics type and value.
//  For readed data returns models.Metrics with value.
//  If type is not gauge or counter, then 501 error.
//  If data is not int64 or float64 then error.
//  If all OK then 200.
func (h *Handlers) HandlePostJSONValue(w http.ResponseWriter, r *http.Request) {
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
		if d, err := h.Repo.GetMetric(*data); err == nil {
			if h.cryptoService.IsEnable {
				_, err = h.cryptoService.Hash(&d)
				if err != nil {
					log.Println("Hasher error!")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
			jData, _ := json.Marshal(d)
			w.WriteHeader(http.StatusOK)
			w.Write(jData)
			return
		} else {
			_, err = h.cryptoService.Hash(&d)
			if err != nil {
				log.Println("Hasher error!")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			jData, _ := json.Marshal(d)
			w.WriteHeader(http.StatusNotFound)
			w.Write(jData)
			return

		}
	}
}

// HandleGetHome returns all data from storage in []models.Storage JSON.
func (h *Handlers) HandleGetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	repo := h.Repo.GetAll(ctx)
	allData, _ := json.MarshalIndent(repo, "", "    ")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(allData))
}

// HandleGetUpdate returns models.Metrics{} fro, URI params.
func (h *Handlers) HandleGetUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	data := models.Metrics{}
	data.ID = nameVal
	data.MType = typeVal
	if valueVar, ok := h.Repo.GetMetric(data); ok != nil {
		http.Error(w, "Name not found!", http.StatusNotFound)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		if valueVar.Value == nil {
			tmp := fmt.Sprintf("%d", *valueVar.Delta)
			w.Write([]byte(tmp))
		} else {
			tmp := strconv.FormatFloat(*valueVar.Value, 'f', -1, 64)
			w.Write([]byte(tmp))
		}

	}

}

// HandlePing return 200 if DB connected.
func (h *Handlers) HandleGetPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if !h.Repo.PingDB() {
		http.Error(w, "Cant connect to DB", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandlePostJSONUpdates get []models.Metrics{} from POST data and batch update it on storage.
func (h *Handlers) HandlePostJSONUpdates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Header.Get("Content-Type") == "application/json" {
		var data models.MetricsDB
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(content, &data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if h.cryptoService.IsEnable {
			for k := range data {
				if !h.cryptoService.CheckHash(data[k]) {
					log.Println("Sing check fail!")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		err = h.Repo.BatchInsert(ctx, data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		commit := models.NewMetric("RandomValue", "gauge", nil, nil, "")
		a := h.Repo.GetCurrentCommit()
		commit.Value = &a
		w.WriteHeader(http.StatusOK)
		jData, _ := json.Marshal(commit)
		w.Write(jData)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		return
	}

}
