package tracegin

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestGinTrackerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	sdk := NewSDK("trace.log", "trace.db")
	defer sdk.Close()

	r.Use(sdk.GinTrackerMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/db-error", func(c *gin.Context) {
		_, err := sql.Open("mysql", "invalid-dsn")
		if err != nil {
			c.Error(err) // Add the error to the context
			c.JSON(500, gin.H{
				"message": "db error",
			})
			return
		}
	})

	r.GET("/logical-error", func(c *gin.Context) {
		err := errors.New("logical error occurred")
		c.Error(err) // Add the error to the context
		c.JSON(500, gin.H{
			"message": "logical error",
		})
	})

	r.GET("/panic", func(c *gin.Context) {
		panic("intentional panic")
	})

	// Test successful request
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "pong"}`, w.Body.String())

	// Test database connection error
	req, _ = http.NewRequest(http.MethodGet, "/db-error", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message": "db error"}`, w.Body.String())

	// Test logical error
	req, _ = http.NewRequest(http.MethodGet, "/logical-error", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message": "logical error"}`, w.Body.String())

	// Test panic error
	req, _ = http.NewRequest(http.MethodGet, "/panic", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"message": "internal server error"}`, w.Body.String())
}
