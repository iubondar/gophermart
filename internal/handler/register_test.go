package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRegisterHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test cases
	tests := []struct {
		name           string
		method         string
		body           []byte
		ucUserID       uuid.UUID
		ucOk           bool
		ucError        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Successful registration",
			method:         http.MethodPost,
			body:           mustMarshal(t, usecase.RegisterIn{Login: "testuser", Password: "testpass"}),
			ucUserID:       uuid.New(),
			ucOk:           true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User already exists",
			method:         http.MethodPost,
			body:           mustMarshal(t, usecase.RegisterIn{Login: "testuser", Password: "testpass"}),
			ucOk:           false,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Usecase error",
			method:         http.MethodPost,
			body:           mustMarshal(t, usecase.RegisterIn{Login: "testuser", Password: "testpass"}),
			ucError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to register user\n",
		},
		{
			name:           "Wrong HTTP method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Only POST requests are allowed!\n",
		},
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			body:           []byte("invalid json"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock usecase
			mockUc := mocks.NewMockRegisterUsecase(ctrl)
			if tt.method == http.MethodPost && tt.expectedStatus != http.StatusBadRequest {
				mockUc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(tt.ucUserID, tt.ucOk, tt.ucError)
			}

			// Create handler
			handler := NewRegisterHandler(mockUc)

			// Create request
			var req *http.Request
			if tt.method == http.MethodPost {
				req = httptest.NewRequest(tt.method, "/api/user/register", bytes.NewBuffer(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/register", nil)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Register(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}

			// Check auth cookie for successful registration
			if tt.expectedStatus == http.StatusOK {
				cookies := resp.Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, auth.AuthCookieName, cookies[0].Name)
			}
		})
	}
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	assert.NoError(t, err)
	return data
}
