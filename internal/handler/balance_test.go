package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/mocks"
	"github.com/iubondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBalanceHandler_Balance(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		userID         uuid.UUID
		account        models.Account
		repoError      error
		expectedStatus int
		expectedBody   BalanceOut
	}{
		{
			name:   "successful balance retrieval",
			method: http.MethodGet,
			userID: uuid.New(),
			account: models.Account{
				Balance:       100.5,
				WithdrawalSum: 50.25,
			},
			expectedStatus: http.StatusOK,
			expectedBody: BalanceOut{
				Current:   100.5,
				Withdrawn: 50.25,
			},
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "repository error",
			method:         http.MethodGet,
			userID:         uuid.New(),
			repoError:      assert.AnError,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockBalanceRepository(ctrl)
			handler := NewBalanceHandler(mockRepo)

			req := httptest.NewRequest(tt.method, "/api/user/balance", nil)
			if tt.userID != uuid.Nil {
				token, err := auth.BuildJWTString(tt.userID)
				require.NoError(t, err)
				req.AddCookie(&http.Cookie{
					Name:  auth.AuthCookieName,
					Value: token,
				})
			}

			if tt.method == http.MethodGet && tt.userID != uuid.Nil {
				mockRepo.EXPECT().
					Account(gomock.Any(), tt.userID).
					Return(tt.account, tt.repoError)
			}

			recorder := httptest.NewRecorder()
			handler.Balance(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK {
				var response BalanceOut
				err := json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}
