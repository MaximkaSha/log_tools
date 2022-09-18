// Agent module collect runtime metrics and send it to remote server.
// Moodule is controled by enviroment variables or console keys.
// All settings are provided in console output.
package main

import (
	"github.com/MaximkaSha/log_tools/internal/agent"
)

func main() {
	agentService := agent.NewAgent()
	agentService.StartService()
}
