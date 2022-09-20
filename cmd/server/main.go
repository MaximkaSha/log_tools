package main

import (
	"fmt"

	"github.com/MaximkaSha/log_tools/internal/server"
)

var (
	BuildVersion string = "N/A"
	BuildTime    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", BuildVersion)
	fmt.Printf("Build date: %s\n", BuildTime)
	fmt.Printf("Build commit: %s\n", BuildCommit)
	var server = server.NewServer()
	server.StartServe()

}
