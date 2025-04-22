package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
)

type Withdrawer interface {
	Withdraw(ctx context.Context, userID uuid.UUID, orderNumber string, sum float32) (result constants.WithdrawResult, err error)
}
