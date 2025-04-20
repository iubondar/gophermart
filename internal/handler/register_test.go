package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegisterHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           RegisterIn
		registrarMock  func(*mocks.MockUserRegistrar)
		expectedStatus int
	}{
		{
			name:   "successful registration",
			method: http.MethodPost,
			body: RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			registrarMock: func(m *mocks.MockUserRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any(), "testuser", gomock.Any()).
					Return(true, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user already exists",
			method: http.MethodPost,
			body: RegisterIn{
				Login:    "existinguser",
				Password: "testpass",
			},
			registrarMock: func(m *mocks.MockUserRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any(), "existinguser", gomock.Any()).
					Return(false, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "registrar error",
			method: http.MethodPost,
			body: RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			registrarMock: func(m *mocks.MockUserRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any(), "testuser", gomock.Any()).
					Return(false, errors.New("database error"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			body: RegisterIn{
				Login:    "testuser",
				Password: "testpass",
			},
			registrarMock:  func(m *mocks.MockUserRegistrar) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "empty login",
			method: http.MethodPost,
			body: RegisterIn{
				Login:    "",
				Password: "testpass",
			},
			registrarMock:  func(m *mocks.MockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty password",
			method: http.MethodPost,
			body: RegisterIn{
				Login:    "testuser",
				Password: "",
			},
			registrarMock:  func(m *mocks.MockUserRegistrar) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mock
			mockRegistrar := mocks.NewMockUserRegistrar(ctrl)
			tt.registrarMock(mockRegistrar)

			// Create handler
			handler := NewRegisterHandler(mockRegistrar)

			// Create request body
			body, err := json.Marshal(tt.body)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(tt.method, "/", bytes.NewBuffer(body))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.Register(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// If successful registration, check for auth cookie
			if tt.expectedStatus == http.StatusOK {
				resp := rr.Result()
				defer resp.Body.Close()
				cookies := resp.Cookies()
				assert.NotEmpty(t, cookies)
				authCookie := cookies[0]
				assert.Equal(t, auth.AuthCookieName, authCookie.Name)
				assert.NotEmpty(t, authCookie.Value)
			}
		})
	}
}

func TestRegisterHandler_Register_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockRegistrar := mocks.NewMockUserRegistrar(ctrl)

	// Create handler
	handler := NewRegisterHandler(mockRegistrar)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("invalid json"))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterHandler_Register_ReadBodyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockRegistrar := mocks.NewMockUserRegistrar(ctrl)

	// Create handler
	handler := NewRegisterHandler(mockRegistrar)

	// Create request with a body that will cause an error when reading
	req := httptest.NewRequest(http.MethodPost, "/", &testhelpers.ErrorReader{})

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.Register(rr, req)

	// Check status code
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
