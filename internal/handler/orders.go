package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/auth"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/models"
)

type OrdersRepository interface {
	Orders(ctx context.Context, userID uuid.UUID) (orders []models.Order, err error)
}

type OrdersOut struct {
	Number     string                `json:"number"`
	Status     constants.OrderStatus `json:"status"`
	Accrual    int                   `json:"accrual"`
	UploadedAt string                `json:"uploaded_at"`
}

type OrdersHandler struct {
	repo OrdersRepository
}

func NewOrdersHandler(repo OrdersRepository) OrdersHandler {
	return OrdersHandler{
		repo: repo,
	}
}

func (handler OrdersHandler) Orders(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.GetUserIDFromReq(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	orders, err := handler.repo.Orders(req.Context(), userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	out := make([]OrdersOut, 0, len(orders))
	for i := range orders {
		outElem := OrdersOut{
			Number:     orders[i].Number,
			Status:     orders[i].Status,
			Accrual:    orders[i].Accrual,
			UploadedAt: orders[i].UploadedAt.Format(time.RFC3339),
		}
		out = append(out, outElem)
	}

	resp, err := json.Marshal(out)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
	} else {
		res.WriteHeader(http.StatusOK)
	}

	res.Write(resp)
}
