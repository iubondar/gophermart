package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
)

type LoginIn struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginChecker interface {
	CheckLogin(ctx context.Context, login string, password_hash string) (userID uuid.UUID, err error)
}

type LoginHandler struct {
	checker LoginChecker
}

func NewLoginHandler(checker LoginChecker) *LoginHandler {
	return &LoginHandler{
		checker: checker,
	}
}

func (handler LoginHandler) Login(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var in LoginIn
	var buf bytes.Buffer
	// читаем тело запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// десериализуем JSON
	if err = json.Unmarshal(buf.Bytes(), &in); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := handler.checker.CheckLogin(req.Context(), in.Login, in.Password)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if userID != uuid.Nil {
		// обновляем авторизационную куку и её срок действия
		err = auth.SetNewAuthCookie(userID, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(http.StatusUnauthorized)
	}
}
