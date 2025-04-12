package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/handler/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrdersHandler_Orders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Test cases
	tests := []struct {
		name           string
		userID         uuid.UUID
		orders         []models.Order
		repoError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful retrieval of orders",
			userID: uuid.New(),
			orders: []models.Order{
				{
					Number:     "1234567890",
					Status:     constants.OrderStatusProcessed,
					Accrual:    100.5,
					UploadedAt: time.Now(),
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No orders found",
			userID:         uuid.New(),
			orders:         []models.Order{},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Repository error",
			userID:         uuid.New(),
			repoError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   assert.AnError.Error() + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock repository
			mockRepo := mocks.NewMockOrdersRepository(ctrl)
			mockRepo.EXPECT().
				Orders(gomock.Any(), tt.userID).
				Return(tt.orders, tt.repoError)

			// Create handler
			handler := NewOrdersHandler(mockRepo)

			// Create request with auth cookie
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			token, err := auth.BuildJWTString(tt.userID)
			assert.NoError(t, err)
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

			// Check response body for successful cases
			if tt.expectedStatus == http.StatusOK {
				var response []OrdersOut
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, len(tt.orders))
				if len(tt.orders) > 0 {
					assert.Equal(t, tt.orders[0].Number, response[0].Number)
					assert.Equal(t, tt.orders[0].Status, response[0].Status)
					assert.Equal(t, tt.orders[0].Accrual, response[0].Accrual)
				}
			} else if tt.expectedStatus == http.StatusBadRequest {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestOrdersHandler_Orders_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repository
	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	handler := NewOrdersHandler(mockRepo)

	// Create request with POST method
	req := httptest.NewRequest(http.MethodPost, "/api/user/orders", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code and response
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	assert.Equal(t, "Only GET requests are allowed!\n", rr.Body.String())
}

func TestOrdersHandler_Orders_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Setup mock repository
	mockRepo := mocks.NewMockOrdersRepository(ctrl)
	handler := NewOrdersHandler(mockRepo)

	// Create request without auth cookie
	req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	rr := httptest.NewRecorder()

	// Call handler
	handler.Orders(rr, req)

	// Check status code and response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Equal(t, "http: named cookie not present\n", rr.Body.String())
}
