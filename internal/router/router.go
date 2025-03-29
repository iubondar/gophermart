package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/gophermart/internal/handler"
	"github.com/iubondar/gophermart/internal/storage"
)

func NewRouter(storage *storage.Storage) (chi.Router, error) {
	router := chi.NewRouter()

	registerHandler := handler.NewRegisterHandler(storage)

	router.Post("/api/user/register", registerHandler.Register)

	return router, nil
}
