package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/usecase"
	"go.uber.org/zap"
)

type RegisterHandler struct {
	uc usecase.RegisterUsecase
}

func NewRegisterHandler(uc usecase.RegisterUsecase) *RegisterHandler {
	return &RegisterHandler{
		uc: uc,
	}
}

func (handler RegisterHandler) Register(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var in usecase.RegisterIn
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

	userID, ok, err := handler.uc.Register(req.Context(), in)
	if err != nil {
		zap.L().Sugar().Debugln("Failed to register user", zap.Error(err))
		http.Error(res, "Failed to register user", http.StatusInternalServerError)
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
