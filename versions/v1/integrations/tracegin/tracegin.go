package tracegin

import (
	"bufio"
	"database/sql"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
	"github.com/sirupsen/logrus"
)

type GinSDK struct {
	Logger     *logrus.Logger
	logFile    *os.File
	bufWriter  *bufio.Writer
	pool       sync.Pool
	shutdownCh chan struct{}
	wg         sync.WaitGroup
	db         *sql.DB
}

type requestInfo struct {
	DateTime      string  `json:"dateTime"`
	RequestMethod string  `json:"requestMethod"`
	RequestURL    string  `json:"requestURL"`
	Status        int     `json:"status"`
	Latency       string  `json:"latency"`
	CPUDelta      float64 `json:"cpuDelta"`
	MemoryDelta   float64 `json:"memoryDelta"`
}

func NewSDK(filename string, dbPath string) *GinSDK {
	logDir := filepath.Dir(filename)
	_ = os.MkdirAll(logDir, 0755)

	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatal(err)
	}

	bufWriter := bufio.NewWriter(logFile)
	logger := logrus.New()
	logger.SetOutput(bufWriter)
	logger.SetFormatter(&logrus.JSONFormatter{})

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logrus.Fatal(err)
	}

	sdk := &GinSDK{
		Logger:     logger,
		logFile:    logFile,
		bufWriter:  bufWriter,
		shutdownCh: make(chan struct{}),
		db:         db,
	}
	sdk.pool.New = func() interface{} {
		return &requestInfo{}
	}
	sdk.wg.Add(1)
	go sdk.periodicFlush()
	sdk.SetupCleanup()
	sdk.initDB()
	return sdk
}

func (sdk *GinSDK) initDB() {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS request_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dateTime TEXT,
		requestMethod TEXT,
		requestURL TEXT,
		status INTEGER,
		latency TEXT,
		cpuDelta REAL,
		memoryDelta REAL
	);`
	_, err := sdk.db.Exec(createTableQuery)
	if err != nil {
		sdk.Logger.Fatal(err)
	}
}

func (sdk *GinSDK) periodicFlush() {
	defer sdk.wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sdk.bufWriter.Flush()
		case <-sdk.shutdownCh:
			return
		}
	}
}

func (sdk *GinSDK) SetupCleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		sdk.Logger.Info("Shutdown signal received, cleaning up...")
		close(sdk.shutdownCh)
		sdk.wg.Wait()
		sdk.bufWriter.Flush()
		sdk.logFile.Close()
		sdk.db.Close()
		sdk.Logger.Info("Cleanup complete")
		os.Exit(0)
	}()
}

func (sdk *GinSDK) Close() {
	close(sdk.shutdownCh)
	sdk.bufWriter.Flush()
	sdk.logFile.Close()
	sdk.db.Close()
}

func (sdk *GinSDK) GinTrackerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		info := sdk.pool.Get().(*requestInfo)
		defer sdk.pool.Put(info)
		start := time.Now()
		proc, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			sdk.Logger.Errorf("Failed to get process: %v", err)
			ctx.Next()
			return
		}
		initialCPUTimes, err := proc.Times()
		if err != nil {
			sdk.Logger.Errorf("Failed to get initial CPU times: %v", err)
			ctx.Next()
			return
		}
		initialMem, err := proc.MemoryInfo()
		if err != nil {
			sdk.Logger.Errorf("Failed to get initial memory info: %v", err)
			return
		}
		ctx.Next()
		// Check for errors that occurred during the request
		defer func() {
			if err := recover(); err != nil {
				sdk.Logger.Errorf("Panic recovered: %v", err)
			}
		}()

		finalCPUTimes, err := proc.Times()
		if err != nil {
			sdk.Logger.Errorf("Failed to get final CPU times: %v", err)
			return
		}
		finalMem, err := proc.MemoryInfo()
		if err != nil {
			sdk.Logger.Errorf("Failed to get final memory info: %v", err)
			return
		}
		cpuUsage := sdk.CalculateCPUUsage(initialCPUTimes, finalCPUTimes, start)
		memoryUsage := int64(finalMem.RSS) - int64(initialMem.RSS)

		info.DateTime = start.Format(time.RFC3339)
		info.RequestMethod = ctx.Request.Method
		info.RequestURL = ctx.Request.URL.Path
		info.Status = ctx.Writer.Status()
		info.Latency = time.Since(start).String()
		info.CPUDelta = cpuUsage
		info.MemoryDelta = float64(memoryUsage) / (1024 * 1024)
		sdk.Logger.WithFields(logrus.Fields{
			"DateTime":          info.DateTime,
			"RequestMethod":     info.RequestMethod,
			"RequestURL":        info.RequestURL,
			"Status":            info.Status,
			"Latency":           info.Latency,
			"CPU Delta":         info.CPUDelta,
			"Memory Delta (MB)": info.MemoryDelta,
		}).Info("Request details logged")

		insertQuery := `
		INSERT INTO request_logs (dateTime, requestMethod, requestURL, status, latency, cpuDelta, memoryDelta)
		VALUES (?, ?, ?, ?, ?, ?, ?);`
		_, err = sdk.db.Exec(insertQuery, info.DateTime, info.RequestMethod, info.RequestURL, info.Status, info.Latency, info.CPUDelta, info.MemoryDelta)
		if err != nil {
			sdk.Logger.Errorf("Failed to insert log into database: %v", err)
		}
	}
}

func (sdk *GinSDK) CalculateCPUUsage(initial, final *cpu.TimesStat, start time.Time) float64 {
	cpuUser := final.User - initial.User
	cpuSystem := final.System - initial.System
	cpuTotal := cpuUser + cpuSystem

	elapsedTime := time.Since(start).Seconds()
	if elapsedTime > 0 {
		return (cpuTotal / elapsedTime) * 100
	}
	return 0
}
