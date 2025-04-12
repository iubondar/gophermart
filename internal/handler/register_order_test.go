package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegisterOrderHandler_RegisterOrder(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name           string
		method         string
		body           string
		registrarMock  func(*mocks.MockOrderRegistrar)
		expectedStatus int
	}{
		{
			name:   "successful registration - already registered",
			method: http.MethodPost,
			body:   "12345678903",
			registrarMock: func(m *mocks.MockOrderRegistrar) {
				m.EXPECT().RegisterOrder(gomock.Any(), userID, "12345678903").
					Return(constants.AlreadyRegistered, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "successful registration - accepted to processing",
			method: http.MethodPost,
			body:   "12345678903",
			registrarMock: func(m *mocks.MockOrderRegistrar) {
				m.EXPECT().RegisterOrder(gomock.Any(), userID, "12345678903").
					Return(constants.AcceptedToProcessing, nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:   "wrong order number format",
			method: http.MethodPost,
			body:   "invalid",
			registrarMock: func(m *mocks.MockOrderRegistrar) {
				m.EXPECT().RegisterOrder(gomock.Any(), userID, "invalid").
					Return(constants.WrongOrderNumberFormat, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "registered by another user",
			method: http.MethodPost,
			body:   "12345678903",
			registrarMock: func(m *mocks.MockOrderRegistrar) {
				m.EXPECT().RegisterOrder(gomock.Any(), userID, "12345678903").
					Return(constants.RegisteredByAnotherUser, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           "12345678903",
			registrarMock:  func(m *mocks.MockOrderRegistrar) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "registrar error",
			method: http.MethodPost,
			body:   "12345678903",
			registrarMock: func(m *mocks.MockOrderRegistrar) {
				m.EXPECT().RegisterOrder(gomock.Any(), userID, "12345678903").
					Return(constants.OrderRegistrationResult(0), errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockRegistrar := mocks.NewMockOrderRegistrar(ctrl)
			tt.registrarMock(mockRegistrar)

			// Create handler
			handler := NewRegisterOrderHandler(mockRegistrar)

			// Create request
			req := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
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
			handler.RegisterOrder(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRegisterOrderHandler_RegisterOrder_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockRegistrar := mocks.NewMockOrderRegistrar(ctrl)

	// Create handler
	handler := NewRegisterOrderHandler(mockRegistrar)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("12345678903"))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.RegisterOrder(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRegisterOrderHandler_RegisterOrder_ReadBodyError(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockRegistrar := mocks.NewMockOrderRegistrar(ctrl)

	// Create handler
	handler := NewRegisterOrderHandler(mockRegistrar)

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
	handler.RegisterOrder(rr, req)

	// Check status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
