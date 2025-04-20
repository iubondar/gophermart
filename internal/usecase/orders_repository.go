package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/models"
)

type OrdersRepository interface {
	Orders(ctx context.Context, userID uuid.UUID) (orders []models.Order, err error)
	OrdersToUpdate(ctx context.Context, limit int) (orders []models.Order, err error)
	UpdateOrders(ctx context.Context, orders []models.OrderStatus) error
}
