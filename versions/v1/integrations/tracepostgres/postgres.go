package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/CloudTraceAI/cloudtraceai-go/versions/v1/cloudtrace"
)

type PostgresMonitor struct {
	Connection *sql.DB
}

func NewPostgresMonitor(conn *sql.DB) *PostgresMonitor {
	return &PostgresMonitor{Connection: conn}
}

// Register initializes the PostgreSQL monitoring system
func (pm *PostgresMonitor) Register() error {
	// Log the registration of PostgreSQL monitoring
	cloudtrace.LogInfo("Registering PostgreSQL monitoring...")
	cloudtrace.LogInfo("Successfully connected to PostgreSQL.")
	return nil
}

// Track logs specific events related to PostgreSQL, such as query execution
func (pm *PostgresMonitor) Track(event string, data map[string]interface{}) error {
	// Log the event being tracked
	cloudtrace.LogInfo(fmt.Sprintf("Tracking PostgreSQL event: %s", event))

	// Log all event data
	for key, value := range data {
		cloudtrace.LogInfo(fmt.Sprintf("Event Data - %s: %v", key, value))
	}

	// Example of tracking a database query execution event:
	if event == "QueryExecution" {
		query := data["query"].(string)
		executionTime := data["execution_time"].(string)
		status := data["status"].(string)

		cloudtrace.LogInfo(fmt.Sprintf("Query: %s | Execution Time: %s | Status: %s", query, executionTime, status))
	}

	return nil
}

// Monitor checks the health of the PostgreSQL connection and queries
func (pm *PostgresMonitor) Monitor() error {
	// Log the monitoring process for PostgreSQL
	cloudtrace.LogInfo("Monitoring PostgreSQL...")

	// Check if the database connection is still alive
	err := pm.Connection.Ping()
	if err != nil {
		cloudtrace.LogError("PostgreSQL connection is down: " + err.Error())
		return err
	}

	// If connection is healthy, log it
	cloudtrace.LogInfo("PostgreSQL connection is healthy.")

	// Example: Monitoring long-running queries (simulating a query)
	// Here, you could check for slow queries or other metrics like table locks
	query := "SELECT pg_stat_activity WHERE state = 'active';"
	start := time.Now()

	// Simulate running a query
	rows, err := pm.Connection.Query(query)
	if err != nil {

		cloudtrace.LogError("Error running query: " + err.Error())
		return err
	}

	elapsed := time.Since(start).String()
	defer rows.Close()

	// Simulate event data for monitoring purposes
	monitoringData := map[string]interface{}{
		"query":          query,
		"execution_time": elapsed,
		"status":         "success",
	}
	// Track the query execution event
	if err := pm.Track("QueryExecution", monitoringData); err != nil {
		cloudtrace.LogError("Error tracking query execution: " + err.Error())
	}
	return nil
}

func (tdb *PostgresMonitor) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, *cloudtrace.DBQueryTrace) {
	trace := &cloudtrace.DBQueryTrace{
		TraceID:    GenerateTraceID(),
		SpanID:     GenerateSpanID(),
		Query:      query,
		StartTime:  time.Now(),
		Attributes: map[string]string{"query_type": "SELECT"},
	}

	rows, err := tdb.Connection.QueryContext(ctx, query, args...)
	trace.EndTime = time.Now()
	trace.Duration = trace.EndTime.Sub(trace.StartTime)

	if err != nil {
		trace.Error = err.Error()
		log.Printf("Error executing query: %v", err)
	}

	// logQuery(trace)
	return rows, trace
}

func (tdb *PostgresMonitor) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, *cloudtrace.DBQueryTrace) {
	trace := &cloudtrace.DBQueryTrace{
		TraceID:    GenerateTraceID(),
		SpanID:     GenerateSpanID(),
		Query:      query,
		StartTime:  time.Now(),
		Attributes: map[string]string{"query_type": "EXEC"},
	}

	result, err := tdb.Connection.ExecContext(ctx, query, args...)
	trace.EndTime = time.Now()
	trace.Duration = trace.EndTime.Sub(trace.StartTime)

	if err != nil {
		trace.Error = err.Error()
		log.Printf("Error executing query: %v", err)
	}

	// logQuery(trace)
	return result, trace
}
