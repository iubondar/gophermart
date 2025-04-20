package usecase

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Register(ctx context.Context, userID uuid.UUID, login string, passwordHash string) (ok bool, err error)
}

type RegisterIn struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterUsecase interface {
	Register(ctx context.Context, in RegisterIn) (userID uuid.UUID, ok bool, err error)
}

type registerUsecase struct {
	repo UserRepository
}

func NewRegisterUsecase(repo UserRepository) RegisterUsecase {
	return &registerUsecase{
		repo: repo,
	}
}

func (uc *registerUsecase) Register(ctx context.Context, in RegisterIn) (userID uuid.UUID, ok bool, err error) {
	if len(in.Login) < 1 || len(in.Password) < 1 {
		return uuid.Nil, false, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, false, err
	}

	userID = uuid.New()
	ok, err = uc.repo.Register(ctx, userID, in.Login, string(hashedPassword))
	if err != nil {
		return uuid.Nil, false, err
	}

	return userID, ok, nil
}
