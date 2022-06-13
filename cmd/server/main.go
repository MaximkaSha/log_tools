package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/caarlos0/env"
	"github.com/go-chi/chi/v5"
)

type Config struct {
	Server string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	repo := storage.NewRepo()
	handl := handlers.NewHandlers(repo)
	mux := chi.NewRouter()
	mux.Post("/update/{type}/{name}/{value}", handl.HandleUpdate)
	mux.Get("/value/{type}/{name}", handl.HandleGetUpdate)
	mux.Get("/", handl.HandleGetHome)
	mux.Post("/update/", handl.HandlePostJSONUpdate)
	mux.Post("/value/", handl.HandlePostJSONValue)

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(cfg.Server, mux))

}
