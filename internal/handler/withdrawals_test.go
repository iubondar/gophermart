package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/usecase"
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
		ucMock         func(*mocks.MockWithdrawalsUsecase)
		expectedStatus int
		expectedBody   []usecase.WithdrawalOut
	}{
		{
			name:   "successful retrieval with withdrawals",
			method: http.MethodGet,
			ucMock: func(m *mocks.MockWithdrawalsUsecase) {
				m.EXPECT().GetWithdrawals(gomock.Any(), userID).
					Return([]usecase.WithdrawalOut{
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
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []usecase.WithdrawalOut{
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
			ucMock: func(m *mocks.MockWithdrawalsUsecase) {
				m.EXPECT().GetWithdrawals(gomock.Any(), userID).
					Return([]usecase.WithdrawalOut{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "usecase error",
			method: http.MethodGet,
			ucMock: func(m *mocks.MockWithdrawalsUsecase) {
				m.EXPECT().GetWithdrawals(gomock.Any(), userID).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			ucMock:         func(m *mocks.MockWithdrawalsUsecase) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockUC := mocks.NewMockWithdrawalsUsecase(ctrl)
			tt.ucMock(mockUC)

			// Create handler
			handler := NewWithdrawalsHandler(mockUC)

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
			if tt.expectedStatus == http.StatusOK {
				var actualBody []usecase.WithdrawalOut
				err := json.Unmarshal(rr.Body.Bytes(), &actualBody)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, actualBody)
			} else if tt.expectedStatus == http.StatusNoContent {
				assert.Empty(t, rr.Body.String())
			}
		})
	}
}

func TestWithdrawalsHandler_Withdrawals_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockWithdrawalsUsecase(ctrl)

	// Create handler
	handler := NewWithdrawalsHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Withdrawals(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
