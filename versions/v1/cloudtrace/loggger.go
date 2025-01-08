package cloudtrace

import (
	"log"
	"os"
	"time"
)

type DBQueryTrace struct {
	TraceID    string            // Unique ID for the trace
	SpanID     string            // Unique ID for the span
	Query      string            // SQL query executed
	StartTime  time.Time         // Query start time
	EndTime    time.Time         // Query end time
	Duration   time.Duration     // Query execution time
	Error      string            // Error message, if any
	Attributes map[string]string // Additional attributes (e.g., database name, table name)
}

var logger *log.Logger

func InitLogger() {
	file, err := os.OpenFile("apm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logger = log.New(file, "APM: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogInfo(message string) {
	logger.Println("INFO: " + message)
}

func LogError(message string) {
	logger.Println("ERROR: " + message)
}

func GenerateTraceID() string {
	return "trace-" + time.Now().Format("20060102150405") // Simplified
}

func GenerateSpanID() string {
	return "span-" + time.Now().Format("150405.000") // Simplified
}
