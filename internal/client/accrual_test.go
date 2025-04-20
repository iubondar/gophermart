package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccrualClient(t *testing.T) {
	address := "http://test-accrual:8080"
	client := NewAccrualClient(address)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpc)
	assert.Equal(t, address, client.httpc.BaseURL)
}

func TestFetchOrderStatus(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name            string
		order           models.Order
		responseStatus  int
		responseBody    string
		expectedStatus  constants.OrderStatus
		expectedAccrual float32
		expectedError   bool
	}{
		{
			name: "successful processing status",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus:  http.StatusOK,
			responseBody:    `{"order": "1234567890", "status": "PROCESSING", "accrual": 100.5}`,
			expectedStatus:  constants.OrderStatusProcessing,
			expectedAccrual: 100.5,
			expectedError:   false,
		},
		{
			name: "successful processed status",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus:  http.StatusOK,
			responseBody:    `{"order": "1234567890", "status": "PROCESSED", "accrual": 200.0}`,
			expectedStatus:  constants.OrderStatusProcessed,
			expectedAccrual: 200.0,
			expectedError:   false,
		},
		{
			name: "invalid order status",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus:  http.StatusOK,
			responseBody:    `{"order": "1234567890", "status": "INVALID", "accrual": 0}`,
			expectedStatus:  constants.OrderStatusInvalid,
			expectedAccrual: 0,
			expectedError:   false,
		},
		{
			name: "no content response",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus:  http.StatusNoContent,
			responseBody:    "",
			expectedStatus:  constants.OrderStatusNew,
			expectedAccrual: 0,
			expectedError:   false,
		},
		{
			name: "server error",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus: http.StatusInternalServerError,
			responseBody:   "",
			expectedError:  true,
		},
		{
			name: "malformed response",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus: http.StatusOK,
			responseBody:   `{"invalid": "json"`,
			expectedError:  true,
		},
		{
			name: "order number mismatch",
			order: models.Order{
				UserID: userID,
				Number: "1234567890",
			},
			responseStatus: http.StatusOK,
			responseBody:   `{"order": "0987654321", "status": "PROCESSED", "accrual": 200.0}`,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request path
				expectedPath := "/api/orders/" + tt.order.Number
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			// Create client with test server URL
			client := NewAccrualClient(server.URL)

			// Call the method
			status, err := client.FetchOrderStatus(context.Background(), tt.order)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.order.UserID, status.UserID)
			assert.Equal(t, tt.order.Number, status.Number)
			assert.Equal(t, tt.expectedStatus, status.Status)
			assert.Equal(t, tt.expectedAccrual, status.Accrual)
		})
	}
}

func TestMapAccrualStatus(t *testing.T) {
	tests := []struct {
		name           string
		accrualStatus  string
		expectedStatus constants.OrderStatus
	}{
		{
			name:           "registered status",
			accrualStatus:  "REGISTERED",
			expectedStatus: constants.OrderStatusProcessing,
		},
		{
			name:           "processing status",
			accrualStatus:  "PROCESSING",
			expectedStatus: constants.OrderStatusProcessing,
		},
		{
			name:           "processed status",
			accrualStatus:  "PROCESSED",
			expectedStatus: constants.OrderStatusProcessed,
		},
		{
			name:           "invalid status",
			accrualStatus:  "INVALID",
			expectedStatus: constants.OrderStatusInvalid,
		},
		{
			name:           "unknown status",
			accrualStatus:  "UNKNOWN",
			expectedStatus: constants.OrderStatusNew,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapAccrualStatus(tt.accrualStatus)
			assert.Equal(t, tt.expectedStatus, result)
		})
	}
}
