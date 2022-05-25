package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	repo := storage.NewRepo()
	handl := handlers.NewHandlers(repo)
	mux := chi.NewRouter()
	mux.Post("/update/{type}/{name}/{value}", handl.HandleUpdate)
	mux.Get("/update/{type}/{name}", handl.HandleGetUpdate)
	mux.Get("/", handl.HandleGetHome)

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))

}
