package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
)

type OrdersOut struct {
	Number     string                `json:"number"`
	Status     constants.OrderStatus `json:"status"`
	Accrual    float32               `json:"accrual"`
	UploadedAt string                `json:"uploaded_at"`
}

type OrdersUsecase interface {
	GetOrders(ctx context.Context, userID uuid.UUID) ([]OrdersOut, error)
}

type ordersUsecase struct {
	repo OrdersRepository
}

func NewOrdersUsecase(repo OrdersRepository) OrdersUsecase {
	return &ordersUsecase{
		repo: repo,
	}
}

func (uc *ordersUsecase) GetOrders(ctx context.Context, userID uuid.UUID) ([]OrdersOut, error) {
	orders, err := uc.repo.Orders(ctx, userID)
	if err != nil {
		return nil, err
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

	return out, nil
}
