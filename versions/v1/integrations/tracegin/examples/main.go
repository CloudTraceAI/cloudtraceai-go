package main

import (
	"github.com/CloudTraceAI/cloudtraceai-go/versions/v1/integrations/tracegin"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	sdk := tracegin.NewSDK("trace.log", "trace.db")
	defer sdk.Close()

	r.Use(sdk.GinTrackerMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
