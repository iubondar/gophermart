package storage

import "github.com/google/uuid"

type Storage struct {
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Register(userID uuid.UUID, login string, password string) (ok bool, err error) {
	// TODO: implementation
	return true, nil
}
