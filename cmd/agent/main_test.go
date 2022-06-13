package main

import (
	//"runtime"
	"runtime"
	"testing"
	"time"
	//	"github.com/MaximkaSha/log_tools/internal/agent"
)

func Test_sendLogs(t *testing.T) {
	type args struct {
		//	ld agent.Agent.logDB,
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//	sendLogs(tt.args.ld)
		})
	}
}

func Test_collectLogs(t *testing.T) {
	type args struct {
		//	ld  *logData
		rtm runtime.MemStats
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "positive",
			args: args{
				//		ld:  new(logData),
				rtm: runtime.MemStats{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//	oldLogs := collectLogs(tt.args.ld, tt.args.rtm)
			time.Sleep(1 * time.Second)
			//	newLogs := collectLogs(tt.args.ld, tt.args.rtm)
			//	if oldLogs == newLogs {
			//		t.Errorf("Logs not collecting! oldRndVal = %x newRndVal =%x", oldLogs, newLogs)
			//	}
		})
	}
}
