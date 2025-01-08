package main

import (
	"database/sql"
	"log"

	"github.com/CloudTraceAI/cloudtraceai-go/versions/v1/cloudtrace"
	"github.com/CloudTraceAI/cloudtraceai-go/versions/v1/integrations/postgres"
)

func main() {
	// Create the service registry
	registry := cloudtrace.NewServiceRegistry()

	conn, err := sql.Open("postgres", "psql://user:password@localhost:5432/dbname")
	if err != nil {
		log.Fatal("Error connecting to PostgreSQL:", err)
	}
	defer conn.Close()

	// Create a PostgresMonitor instance
	postgresMonitor := postgres.NewPostgresMonitor(conn)
	registry.AddService(postgresMonitor)
	// Initialize all services
	if err := registry.Initialize(); err != nil {
		cloudtrace.LogError("Initialization failed: " + err.Error())
		panic(err)
	}

	// Start monitoring all services
	if err := registry.StartMonitoring(); err != nil {
		cloudtrace.LogError("Monitoring failed: " + err.Error())
		panic(err)
	}

	// Example of tracking an event
	eventData := map[string]interface{}{
		"query":          "SELECT * FROM users WHERE active = TRUE",
		"execution_time": "120ms",
		"status":         "success",
	}
	if err := registry.TrackEvent("QueryExecution", eventData); err != nil {
		cloudtrace.LogError("Event tracking failed: " + err.Error())
	}

	// You can simulate monitoring of PostgreSQL health
	if err := postgresMonitor.Monitor(); err != nil {
		cloudtrace.LogError("PostgreSQL monitoring failed: " + err.Error())
	}
}
