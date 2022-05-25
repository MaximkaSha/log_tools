package handlers

import (
	"net/http"
	"strings"

	"github.com/MaximkaSha/log_tools/internal/storage"
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
	defer r.Body.Close()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}
	tdv := strings.Split(r.RequestURI, "/")[2:]
	if len(tdv) != 3 {
		http.Error(w, "Name or value not found!", http.StatusNotFound)
		return
	}
	if (tdv[0] != "gauge") && (tdv[0] != "counter") {
		http.Error(w, "Type not found!", http.StatusNotImplemented)
		return
	}
	result := obj.repo.InsertData(tdv[0], tdv[1], tdv[2])
	if result != 200 {
		http.Error(w, "Bad value found!", result)
		return
	}
	w.WriteHeader(http.StatusOK)

}
