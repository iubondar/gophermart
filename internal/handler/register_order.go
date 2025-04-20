package handler

import (
	"context"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
)

type OrderRegistrar interface {
	RegisterOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (result constants.OrderRegistrationResult, err error)
}

type RegisterOrderHandler struct {
	registrar OrderRegistrar
}

func NewRegisterOrderHandler(registrar OrderRegistrar) *RegisterOrderHandler {
	return &RegisterOrderHandler{
		registrar: registrar,
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

	result, err := handler.registrar.RegisterOrder(req.Context(), userID, string(body))
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
