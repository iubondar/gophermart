package usecase

import (
	"context"

	"github.com/google/uuid"
)

type LoginRepository interface {
	CheckLogin(ctx context.Context, login string, password string) (userID uuid.UUID, err error)
}

type LoginUsecase interface {
	Login(ctx context.Context, login string, password string) (userID uuid.UUID, err error)
}

type loginUsecase struct {
	repo LoginRepository
}

func NewLoginUsecase(repo LoginRepository) LoginUsecase {
	return &loginUsecase{
		repo: repo,
	}
}

func (uc *loginUsecase) Login(ctx context.Context, login string, password string) (userID uuid.UUID, err error) {
	return uc.repo.CheckLogin(ctx, login, password)
}
