package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/handler/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWithdrawalsHandler_Withdrawals(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name           string
		method         string
		repoMock       func(*mocks.MockWithdrawalsRepository)
		expectedStatus int
		expectedBody   []WithdrawalsOut
	}{
		{
			name:   "successful retrieval with withdrawals",
			method: http.MethodGet,
			repoMock: func(m *mocks.MockWithdrawalsRepository) {
				m.EXPECT().Withdrawals(gomock.Any(), userID).
					Return([]models.Withdrawal{
						{
							Number:      "12345678903",
							Sum:         100.50,
							ProcessedAt: now,
						},
						{
							Number:      "98765432109",
							Sum:         200.75,
							ProcessedAt: now.Add(time.Hour),
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []WithdrawalsOut{
				{
					Order:       "12345678903",
					Sum:         100.50,
					ProcessedAt: now.Format(time.RFC3339),
				},
				{
					Order:       "98765432109",
					Sum:         200.75,
					ProcessedAt: now.Add(time.Hour).Format(time.RFC3339),
				},
			},
		},
		{
			name:   "successful retrieval with no withdrawals",
			method: http.MethodGet,
			repoMock: func(m *mocks.MockWithdrawalsRepository) {
				m.EXPECT().Withdrawals(gomock.Any(), userID).
					Return([]models.Withdrawal{}, nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   []WithdrawalsOut{},
		},
		{
			name:   "repository error",
			method: http.MethodGet,
			repoMock: func(m *mocks.MockWithdrawalsRepository) {
				m.EXPECT().Withdrawals(gomock.Any(), userID).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			repoMock:       func(m *mocks.MockWithdrawalsRepository) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockRepo := mocks.NewMockWithdrawalsRepository(ctrl)
			tt.repoMock(mockRepo)

			// Create handler
			handler := NewWithdrawalsHandler(mockRepo)

			// Create request
			req := httptest.NewRequest(tt.method, "/", nil)
			req = req.WithContext(ctx)

			// Set up auth cookie
			token, err := auth.BuildJWTString(userID)
			require.NoError(t, err)
			req.AddCookie(&http.Cookie{
				Name:  auth.AuthCookieName,
				Value: token,
			})

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Withdrawals(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if expected
			if tt.expectedBody != nil {
				var actualBody []WithdrawalsOut
				err := json.Unmarshal(rr.Body.Bytes(), &actualBody)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, actualBody)
			}
		})
	}
}

func TestWithdrawalsHandler_Withdrawals_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockRepo := mocks.NewMockWithdrawalsRepository(ctrl)

	// Create handler
	handler := NewWithdrawalsHandler(mockRepo)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Withdrawals(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
