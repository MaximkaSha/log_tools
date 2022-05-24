package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MaximkaSha/log_tools/internal/storage"
	"github.com/MaximkaSha/log_tools/internal/utils"
)

func main() {
	repo := storage.NewRepo()
	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Type not found!", http.StatusNotFound)
			return
		}
		if tdv[0] == "gauge" {
			if utils.CheckIfStringIsNumber(tdv[2]) {
				repo.InsertGouge(tdv[1], tdv[2])
			} else {
				http.Error(w, "Bad value found!", http.StatusBadRequest)
				return
			}
		}
		if tdv[0] == "counter" {
			if utils.CheckIfStringIsNumber(tdv[2]) {
				repo.InsertCount(tdv[1], tdv[2])
			} else {
				http.Error(w, "Bad value found!", http.StatusBadRequest)
				return
			}
		}
		fmt.Print(repo)
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Server is listening...")
	http.ListenAndServe("127.0.0.1:8080", nil)

}
