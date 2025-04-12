package logging

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoggingResponseWriter_Write(t *testing.T) {
	// Create a test response writer
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	// Test writing data
	data := []byte("test data")
	size, err := lw.Write(data)

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, len(data), size)
	assert.Equal(t, len(data), responseData.size)
	assert.Equal(t, data, recorder.Body.Bytes())
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	// Create a test response writer
	recorder := httptest.NewRecorder()
	responseData := &responseData{}
	lw := loggingResponseWriter{
		ResponseWriter: recorder,
		responseData:   responseData,
	}

	// Test setting status code
	statusCode := http.StatusNotFound
	lw.WriteHeader(statusCode)

	// Verify results
	assert.Equal(t, statusCode, responseData.status)
	assert.Equal(t, statusCode, recorder.Code)
}

func TestWithLogging(t *testing.T) {
	// Create a test handler that writes some data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	recorder := httptest.NewRecorder()

	// Wrap the handler with logging middleware
	loggingHandler := WithLogging(handler)

	// Record start time
	start := time.Now()

	// Serve the request
	loggingHandler.ServeHTTP(recorder, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "test response", recorder.Body.String())

	// Verify that the request took some time (not zero)
	assert.True(t, time.Since(start) > 0)
}
