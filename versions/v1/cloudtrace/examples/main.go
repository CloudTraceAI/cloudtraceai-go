package main

import (
	"fmt"

	cloudtrace "github.com/CloudTraceAI/cloudtraceai-go/versions/v1/cloudtrace"
)

func main() {
	registry := cloudtrace.NewServiceRegistry()
	monitor := cloudtrace.NewCodeMonitor("MyApp")
	registry.AddService(monitor)
	// cloudtrace.DbqueryTrace("SELECT * FROM users WHERE active = TRUE", "120ms",)
	monitor.Track("App started", map[string]interface{}{"version": "1.0.0"})
	fmt.Println("Application registered successfully with Monitoring Package")
}
