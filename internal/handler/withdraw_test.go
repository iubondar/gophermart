package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/handler"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/usecase"
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
		body           usecase.WithdrawIn
		ucMock         func(*mocks.MockWithdrawUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful withdrawal",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   100.50,
				}).Return(constants.Success, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "insufficient funds",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   100.50,
				}).Return(constants.InsufficientFunds, nil)
			},
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:   "wrong order format",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   100.50,
				}).Return(constants.WrongOrderFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "empty order number",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "",
				Sum:   100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "",
					Sum:   100.50,
				}).Return(constants.WrongOrderFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "zero sum",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   0,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   0,
				}).Return(constants.WrongOrderFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "negative sum",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   -100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   -100.50,
				}).Return(constants.WrongOrderFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "internal error",
			method: http.MethodPost,
			body: usecase.WithdrawIn{
				Order: "12345678903",
				Sum:   100.50,
			},
			ucMock: func(m *mocks.MockWithdrawUsecase) {
				m.EXPECT().Withdraw(gomock.Any(), userID, usecase.WithdrawIn{
					Order: "12345678903",
					Sum:   100.50,
				}).Return(constants.Success, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           usecase.WithdrawIn{},
			ucMock:         func(m *mocks.MockWithdrawUsecase) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Only POST requests are allowed!\n",
		},
		{
			name:           "invalid json",
			method:         http.MethodPost,
			body:           usecase.WithdrawIn{},
			ucMock:         func(m *mocks.MockWithdrawUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockUC := mocks.NewMockWithdrawUsecase(ctrl)
			tt.ucMock(mockUC)

			// Create handler
			handler := handler.NewWithdrawHandler(mockUC)

			// Create request
			var req *http.Request
			if tt.name == "invalid json" {
				req = httptest.NewRequest(tt.method, "/", bytes.NewBufferString("invalid json"))
			} else {
				body, err := json.Marshal(tt.body)
				require.NoError(t, err)
				req = httptest.NewRequest(tt.method, "/", bytes.NewBuffer(body))
			}
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

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestWithdrawHandler_Withdraw_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockWithdrawUsecase(ctrl)

	// Create handler
	handler := handler.NewWithdrawHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Withdraw(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
