// Agent module collect runtime metrics and send it to remote server.
// Moodule is controled by enviroment variables or console keys.
// All settings are provided in console output.
package main

import (
	"fmt"

	"github.com/MaximkaSha/log_tools/internal/agent"
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
	agentService := agent.NewAgent()
	agentService.StartService()

}
