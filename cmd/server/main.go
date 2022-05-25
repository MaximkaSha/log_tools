package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MaximkaSha/log_tools/internal/handlers"
	"github.com/MaximkaSha/log_tools/internal/storage"
)

func main() {
	repo := storage.NewRepo()
	handl := handlers.NewHandlers(repo)
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handl.HandleUpdate)

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))

}
