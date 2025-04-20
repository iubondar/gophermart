package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestOrdersHandler_Orders(t *testing.T) {
	userID := uuid.New()
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name           string
		method         string
		ucMock         func(*mocks.MockOrdersUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful retrieval with orders",
			method: http.MethodGet,
			ucMock: func(m *mocks.MockOrdersUsecase) {
				m.EXPECT().GetOrders(gomock.Any(), userID).Return([]usecase.OrdersOut{
					{
						Number:     "12345678903",
						Status:     constants.OrderStatusProcessed,
						Accrual:    100.50,
						UploadedAt: now.Format(time.RFC3339),
					},
					{
						Number:     "12345678904",
						Status:     constants.OrderStatusProcessing,
						Accrual:    0,
						UploadedAt: now.Format(time.RFC3339),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"number":"12345678903","status":"PROCESSED","accrual":100.5,"uploaded_at":"` + now.Format(time.RFC3339) + `"},{"number":"12345678904","status":"PROCESSING","accrual":0,"uploaded_at":"` + now.Format(time.RFC3339) + `"}]`,
		},
		{
			name:   "successful retrieval with no orders",
			method: http.MethodGet,
			ucMock: func(m *mocks.MockOrdersUsecase) {
				m.EXPECT().GetOrders(gomock.Any(), userID).Return([]usecase.OrdersOut{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "repository error",
			method: http.MethodGet,
			ucMock: func(m *mocks.MockOrdersUsecase) {
				m.EXPECT().GetOrders(gomock.Any(), userID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   assert.AnError.Error() + "\n",
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			ucMock:         func(m *mocks.MockOrdersUsecase) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Only GET requests are allowed!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockUC := mocks.NewMockOrdersUsecase(ctrl)
			tt.ucMock(mockUC)

			// Create handler
			handler := handler.NewOrdersHandler(mockUC)

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
			handler.Orders(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestOrdersHandler_Orders_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "http: named cookie not present\n", rr.Body.String())
}

func TestOrdersHandler_Orders_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestOrdersHandler_Orders_UsecaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestOrdersHandler_Orders_SuccessNoOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestOrdersHandler_Orders_SuccessWithOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOrdersHandler_Orders_SuccessWithOrders_Formatted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body
	expectedBody := `[{"number":"12345678903","status":"PROCESSED","accrual":100.5,"uploaded_at":"` + time.Now().Format(time.RFC3339) + `"},{"number":"12345678904","status":"PROCESSING","accrual":0,"uploaded_at":"` + time.Now().Format(time.RFC3339) + `"}]`
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestOrdersHandler_Orders_SuccessWithOrders_Formatted_NoOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body
	expectedBody := `[]`
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestOrdersHandler_Orders_SuccessWithOrders_Formatted_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// Check response body
	expectedBody := assert.AnError.Error() + "\n"
	assert.Equal(t, expectedBody, rr.Body.String())
}

func TestOrdersHandler_Orders_SuccessWithOrders_Formatted_WrongMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockUC := mocks.NewMockOrdersUsecase(ctrl)

	// Create handler
	handler := handler.NewOrdersHandler(mockUC)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Check response body
	expectedBody := "Only GET requests are allowed!\n"
	assert.Equal(t, expectedBody, rr.Body.String())
}
