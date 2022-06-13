package main

import (
	"github.com/MaximkaSha/log_tools/internal/server"
)

func main() {
	var server = server.NewServer()
	server.StartServe()

}
