package handler

import (
	"io"
	"net/http"

	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/usecase"
)

type RegisterOrderHandler struct {
	uc usecase.RegisterOrderUsecase
}

func NewRegisterOrderHandler(uc usecase.RegisterOrderUsecase) *RegisterOrderHandler {
	return &RegisterOrderHandler{
		uc: uc,
	}
}

func (handler RegisterOrderHandler) RegisterOrder(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := handler.uc.RegisterOrder(req.Context(), userID, string(body))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	switch result {
	case constants.AlreadyRegistered:
		res.WriteHeader(http.StatusOK)
	case constants.AcceptedToProcessing:
		res.WriteHeader(http.StatusAccepted)
	case constants.WrongOrderNumberFormat:
		res.WriteHeader(http.StatusUnprocessableEntity)
	case constants.RegisteredByAnotherUser:
		res.WriteHeader(http.StatusConflict)
	default:
		http.Error(res, "Wrong format", http.StatusBadRequest)
	}
}
