package handler

import (
	"bytes"
	"context"
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test cases
	tests := []struct {
		name           string
		method         string
		userID         uuid.UUID
		orderNumber    string
		ucResult       constants.OrderRegistrationResult
		ucError        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Successful order registration",
			method:         http.MethodPost,
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			ucResult:       constants.AcceptedToProcessing,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Already registered order",
			method:         http.MethodPost,
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			ucResult:       constants.AlreadyRegistered,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Wrong order number format",
			method:         http.MethodPost,
			userID:         uuid.New(),
			orderNumber:    "invalid",
			ucResult:       constants.WrongOrderNumberFormat,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Order registered by another user",
			method:         http.MethodPost,
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			ucResult:       constants.RegisteredByAnotherUser,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Usecase error",
			method:         http.MethodPost,
			userID:         uuid.New(),
			orderNumber:    "1234567890",
			ucError:        assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   assert.AnError.Error() + "\n",
		},
		{
			name:           "Wrong HTTP method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Only POST requests are allowed!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock usecase
			mockUc := mocks.NewMockRegisterOrderUsecase(ctrl)
			if tt.method == http.MethodPost {
				mockUc.EXPECT().
					RegisterOrder(gomock.Any(), tt.userID, tt.orderNumber).
					Return(tt.ucResult, tt.ucError)
			}

			// Create handler
			handler := NewRegisterOrderHandler(mockUc)

			// Create request
			var req *http.Request
			if tt.method == http.MethodPost {
				req = httptest.NewRequest(tt.method, "/api/user/orders", bytes.NewBufferString(tt.orderNumber))
			} else {
				req = httptest.NewRequest(tt.method, "/api/user/orders", nil)
			}

			// Add auth cookie if userID is set
			if tt.userID != uuid.Nil {
				token, err := auth.BuildJWTString(tt.userID)
				assert.NoError(t, err)
				req.AddCookie(&http.Cookie{
					Name:  auth.AuthCookieName,
					Value: token,
				})
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RegisterOrder(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
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
