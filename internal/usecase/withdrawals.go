package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/models"
)

type WithdrawalsRepository interface {
	Withdrawals(ctx context.Context, userID uuid.UUID) (withdrawal []models.Withdrawal, err error)
}

type WithdrawalsUsecase interface {
	GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]WithdrawalOut, error)
}

type WithdrawalOut struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type withdrawalsUsecase struct {
	repo WithdrawalsRepository
}

func NewWithdrawalsUsecase(repo WithdrawalsRepository) WithdrawalsUsecase {
	return &withdrawalsUsecase{
		repo: repo,
	}
}

func (uc *withdrawalsUsecase) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]WithdrawalOut, error) {
	withdrawals, err := uc.repo.Withdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]WithdrawalOut, 0, len(withdrawals))
	for i := range withdrawals {
		outElem := WithdrawalOut{
			Order:       withdrawals[i].Number,
			Sum:         withdrawals[i].Sum,
			ProcessedAt: withdrawals[i].ProcessedAt.Format(time.RFC3339),
		}
		out = append(out, outElem)
	}

	return out, nil
}
