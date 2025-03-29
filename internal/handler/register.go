package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

type RegisterIn struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRegistrator interface {
	Register(ctx context.Context, userID uuid.UUID, login string, password_hash string) (ok bool, err error)
}

type RegisterHandler struct {
	registrator UserRegistrator
}

func NewRegisterHandler(registrator UserRegistrator) *RegisterHandler {
	return &RegisterHandler{
		registrator: registrator,
	}
}

func (handler RegisterHandler) Register(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var in RegisterIn
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

	if len(in.Login) < 1 || len(in.Password) < 1 {
		http.Error(res, "Login or password is empty", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(res, "Error hashing password "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID := uuid.New()
	ok, err := handler.registrator.Register(req.Context(), userID, in.Login, string(hashedPassword))
	if err != nil {
		http.Error(res, "Failed to register user", http.StatusBadRequest)
		return
	}

	if !ok {
		res.WriteHeader(http.StatusConflict)
		return
	}

	err = auth.SetNewAuthCookie(userID, res)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
