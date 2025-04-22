package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
)

type OrderRepository interface {
	RegisterOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (result constants.OrderRegistrationResult, err error)
}

type RegisterOrderUsecase interface {
	RegisterOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (result constants.OrderRegistrationResult, err error)
}

type registerOrderUsecase struct {
	repo OrderRepository
}

func NewRegisterOrderUsecase(repo OrderRepository) RegisterOrderUsecase {
	return &registerOrderUsecase{
		repo: repo,
	}
}

func (uc *registerOrderUsecase) RegisterOrder(ctx context.Context, userID uuid.UUID, orderNumber string) (result constants.OrderRegistrationResult, err error) {
	return uc.repo.RegisterOrder(ctx, userID, orderNumber)
}
