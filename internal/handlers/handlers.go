package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/MaximkaSha/log_tools/internal/crypto"
	"github.com/MaximkaSha/log_tools/internal/database"
	"github.com/MaximkaSha/log_tools/internal/models"

	//"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	handlers      *http.ServeMux
	Repo          models.Storager
	SyncFile      string
	cryptoService crypto.CryptoService
	DB            *database.Database
}

func NewHandlers(repo models.Storager, cryptoService crypto.CryptoService) Handlers {
	handl := http.NewServeMux()
	return Handlers{
		handlers:      handl,
		Repo:          repo,
		SyncFile:      "",
		cryptoService: cryptoService,
	}
}

func (obj *Handlers) HandleUpdate(w http.ResponseWriter, r *http.Request) { //should be renamed to HandlePostUpdate

	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	valueVal := chi.URLParam(r, "value")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if (typeVal != "gauge") && (typeVal != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	var data models.Metrics
	if obj.cryptoService.IsServiceEnable() {
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
		obj.cryptoService.Hash(&data)
	}

	result := obj.Repo.InsertData(typeVal, nameVal, valueVal, data.Hash)
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
		if obj.cryptoService.IsEnable {
			if !obj.cryptoService.CheckHash(*data) {
				log.Println("Sing check fail!")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
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
		if d, err := obj.Repo.GetMetric(*data); err == nil {
			if obj.cryptoService.IsEnable {
				_, err = obj.cryptoService.Hash(&d)
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
			_, err = obj.cryptoService.Hash(&d)
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

func (obj *Handlers) HandleGetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	repo := obj.Repo.GetAll()
	allData, _ := json.MarshalIndent(repo, "", "    ")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(allData))
}

func (obj *Handlers) HandleGetUpdate(w http.ResponseWriter, r *http.Request) {
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
	if valueVar, ok := obj.Repo.GetMetric(data); ok != nil {
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

func (obj *Handlers) HandleGetPing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if !obj.Repo.PingDB() {
		http.Error(w, "Cant connect to DB", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
