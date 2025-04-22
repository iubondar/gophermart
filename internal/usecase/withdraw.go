package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
)

type WithdrawIn struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

type WithdrawUsecase interface {
	Withdraw(ctx context.Context, userID uuid.UUID, in WithdrawIn) (result constants.WithdrawResult, err error)
}

type withdrawUsecase struct {
	withdrawer Withdrawer
}

func NewWithdrawUsecase(withdrawer Withdrawer) WithdrawUsecase {
	return &withdrawUsecase{
		withdrawer: withdrawer,
	}
}

func (uc *withdrawUsecase) Withdraw(ctx context.Context, userID uuid.UUID, in WithdrawIn) (result constants.WithdrawResult, err error) {
	if in.Order == "" {
		return constants.WrongOrderFormat, nil
	}

	if in.Sum <= 0 {
		return constants.WrongOrderFormat, nil
	}

	result, err = uc.withdrawer.Withdraw(ctx, userID, in.Order, in.Sum)
	if err != nil {
		return constants.Success, err
	}
	return result, nil
}
