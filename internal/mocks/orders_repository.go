package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/models"
)

//go:generate mockgen -source=orders_repository.go -destination=./mock_orders_repository.go -package=mocks
type OrdersRepository interface {
	Orders(ctx context.Context, userID uuid.UUID) (orders []models.Order, err error)
	OrdersToUpdate(ctx context.Context, limit int) (orders []models.Order, err error)
	UpdateOrders(ctx context.Context, orders []models.OrderStatus) error
}
