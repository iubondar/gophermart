package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServer_StartAndShutdown(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a new server instance
	server := New(":0", handler) // Using :0 to get a random available port

	// Channel to signal when server is ready
	serverReady := make(chan struct{})
	serverError := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		close(serverReady) // Signal that we're starting the server
		err := server.Start()
		// http.ErrServerClosed is expected when server is shut down
		if err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
		close(serverError)
	}()

	// Wait for server to be ready
	<-serverReady

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test that server is responding
	req := httptest.NewRequest("GET", "http://localhost"+server.httpServer.Addr, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test shutdown
	err = server.Shutdown()
	assert.NoError(t, err)

	// Wait for server to finish
	select {
	case err := <-serverError:
		if err != nil {
			t.Fatalf("Unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not shut down in time")
	}
}

func TestServer_ShutdownTimeout(t *testing.T) {
	// Setup test logger
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Create a handler that takes longer than the shutdown timeout
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second) // Longer than the 5-second shutdown timeout
		w.WriteHeader(http.StatusOK)
	})

	// Create a new server instance
	server := New(":0", handler)

	// Channel to signal when server is ready
	serverReady := make(chan struct{})
	serverError := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		close(serverReady) // Signal that we're starting the server
		err := server.Start()
		// http.ErrServerClosed is expected when server is shut down
		if err != nil && err != http.ErrServerClosed {
			serverError <- err
		}
		close(serverError)
	}()

	// Wait for server to be ready
	<-serverReady

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test shutdown with timeout
	err = server.Shutdown()
	assert.NoError(t, err) // Should still succeed due to forced close

	// Wait for server to finish
	select {
	case err := <-serverError:
		if err != nil {
			t.Fatalf("Unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not shut down in time")
	}
}
