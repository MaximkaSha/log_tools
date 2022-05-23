package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/MaximkaSha/log_tools/internal/storage"
)

func main() {
	var logData = new(storage.LogData)
	http.HandleFunc("/update/gauge/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		if "text/plain" != r.Header.Get("Content-type") {
			http.Error(w, "Only text/plain are allowed!", http.StatusBadGateway)
			return
		}
		s := strings.Split(r.RequestURI, "/")
		reflectLogData := reflect.ValueOf(logData)
		f := reflect.Indirect(reflectLogData).FieldByName(s[3])
		ff, _ := strconv.ParseFloat(s[4], 64)
		f.SetFloat(ff)
		log.Printf("Added data #%s with rnd %s", s[3], s[4])
		http.Error(w, "OK", http.StatusOK)
	})
	http.HandleFunc("/update/counter/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			fmt.Println("post")
			return
		}
		if "text/plain" != r.Header.Get("Content-type") {
			http.Error(w, "Only text/plain are allowed!", http.StatusBadGateway)
			fmt.Println(r.Header.Get("Content-type"))
			return
		}
		s := strings.Split(r.RequestURI, "/")
		reflectLogData := reflect.ValueOf(logData)
		f := reflect.Indirect(reflectLogData).FieldByName(s[3])
		//fmt.Fprint(w, f)
		oldData := f.Int()
		ff, _ := strconv.ParseInt(s[4], 10, 64)
		f.SetInt(oldData + ff)
		log.Printf("Added data %s with rnd %s", s[3], s[4])
		fmt.Println(logData)
		http.Error(w, "OK", http.StatusOK)
	})

	//fmt.Print(logData)
	fmt.Println("Server is listening...")
	http.ListenAndServe("127.0.0.1:8080", nil)

}
