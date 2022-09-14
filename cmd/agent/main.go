package main

import (
	"github.com/MaximkaSha/log_tools/internal/agent"
)

func main() {
	agentService := agent.NewAgent()
	agentService.StartService()
}
