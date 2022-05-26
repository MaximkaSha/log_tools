package handlers

import (
	"fmt"
	"net/http"

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

func (obj Handlers) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	//defer r.Body.Close()
	//fmt.Println("post")
	typeVal := chi.URLParam(r, "type")
	nameVal := chi.URLParam(r, "name")
	valueVal := chi.URLParam(r, "value")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
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

/* Сервер должен возвращать текущее значение запрашиваемой
метрики в текстовом виде по запросу
GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> (со статусом http.StatusOK).
При попытке запроса неизвестной серверу метрики сервер должен возвращать http.StatusNotFound.
По запросу
 GET http://<АДРЕС_СЕРВЕРА>/
 с ервер должен отдавать HTML-страничку со списком имён и значений всех известных ему на текущий момент метрик. */

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
	//fmt.Println("get")
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
