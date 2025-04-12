package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/handler/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		loginIn        LoginIn
		userID         uuid.UUID
		checkerError   error
		expectedStatus int
	}{
		{
			name:   "successful login",
			method: http.MethodPost,
			loginIn: LoginIn{
				Login:    "testuser",
				Password: "testpass",
			},
			userID:         uuid.New(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "invalid json",
			method: http.MethodPost,
			loginIn: LoginIn{
				Login:    "testuser",
				Password: "testpass",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "unauthorized",
			method: http.MethodPost,
			loginIn: LoginIn{
				Login:    "testuser",
				Password: "testpass",
			},
			userID:         uuid.Nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "checker error",
			method: http.MethodPost,
			loginIn: LoginIn{
				Login:    "testuser",
				Password: "testpass",
			},
			checkerError:   assert.AnError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChecker := mocks.NewMockLoginChecker(ctrl)
			handler := NewLoginHandler(mockChecker)

			var reqBody []byte
			if tt.name != "invalid json" {
				var err error
				reqBody, err = json.Marshal(tt.loginIn)
				require.NoError(t, err)
			} else {
				reqBody = []byte("invalid json")
			}

			req := httptest.NewRequest(tt.method, "/api/user/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			if tt.method == http.MethodPost && tt.name != "invalid json" {
				mockChecker.EXPECT().
					CheckLogin(gomock.Any(), tt.loginIn.Login, tt.loginIn.Password).
					Return(tt.userID, tt.checkerError)
			}

			recorder := httptest.NewRecorder()
			handler.Login(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				cookies := recorder.Result().Cookies()
				assert.Len(t, cookies, 1)
				assert.Equal(t, auth.AuthCookieName, cookies[0].Name)
				assert.NotEmpty(t, cookies[0].Value)
			}
		})
	}
}
