package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/models"
)

type BalanceRepository interface {
	Account(ctx context.Context, userID uuid.UUID) (account models.Account, err error)
}

type BalanceOut struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type GetBalanceUsecase interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (BalanceOut, error)
}

type getBalanceUsecase struct {
	repo BalanceRepository
}

func NewGetBalanceUsecase(repo BalanceRepository) GetBalanceUsecase {
	return &getBalanceUsecase{
		repo: repo,
	}
}

func (uc *getBalanceUsecase) GetBalance(ctx context.Context, userID uuid.UUID) (BalanceOut, error) {
	account, err := uc.repo.Account(ctx, userID)
	if err != nil {
		return BalanceOut{}, err
	}

	out := BalanceOut{
		Current:   account.Balance,
		Withdrawn: account.WithdrawalSum,
	}

	return out, nil
}
