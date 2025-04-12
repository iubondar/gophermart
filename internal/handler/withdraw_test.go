package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/handler/mocks"
	"github.com/iubondar/gophermart/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWithdrawHandler_Withdraw(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name           string
		method         string
		body           WithdrawIn
		withdrawerMock func(*mocks.MockWithdrawer)
		expectedStatus int
	}{
		{
			name:   "successful withdrawal",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), userID, "12345678903", float32(100.50)).
					Return(constants.Success, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "insufficient funds",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   1000.00,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), userID, "12345678903", float32(1000.00)).
					Return(constants.InsufficientFunds, nil)
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:   "wrong order format",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "invalid",
				Sum:   100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), userID, "invalid", float32(100.50)).
					Return(constants.WrongOrderFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "withdrawer error",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {
				m.EXPECT().Withdraw(gomock.Any(), userID, "12345678903", float32(100.50)).
					Return(constants.WithdrawResult(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "empty order",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "",
				Sum:   100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "negative sum",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   -100.50,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "zero sum",
			method: http.MethodPost,
			body: WithdrawIn{
				Order: "12345678903",
				Sum:   0,
			},
			withdrawerMock: func(m *mocks.MockWithdrawer) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockWithdrawer := mocks.NewMockWithdrawer(ctrl)
			tt.withdrawerMock(mockWithdrawer)

			// Create handler
			handler := NewWithdrawHandler(mockWithdrawer)

			// Create request body
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(tt.method, "/", bytes.NewBuffer(body))
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
			handler.Withdraw(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestWithdrawHandler_Withdraw_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockWithdrawer := mocks.NewMockWithdrawer(ctrl)

	// Create handler
	handler := NewWithdrawHandler(mockWithdrawer)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"order": "12345678903", "sum": 100.50}`))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Withdraw(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestWithdrawHandler_Withdraw_InvalidJSON(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockWithdrawer := mocks.NewMockWithdrawer(ctrl)

	// Create handler
	handler := NewWithdrawHandler(mockWithdrawer)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid json"))
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
	handler.Withdraw(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestWithdrawHandler_Withdraw_ReadBodyError(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockWithdrawer := mocks.NewMockWithdrawer(ctrl)

	// Create handler
	handler := NewWithdrawHandler(mockWithdrawer)

	// Create request with a body that will cause an error when reading
	req := httptest.NewRequest(http.MethodPost, "/", &testhelpers.ErrorReader{})
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
	handler.Withdraw(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
