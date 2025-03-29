package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/gophermart/internal/handler"
	"github.com/iubondar/gophermart/internal/storage"
)

func NewRouter(storage *storage.Storage) (chi.Router, error) {
	router := chi.NewRouter()

	registerHandler := handler.NewRegisterHandler(storage)
	loginHandler := handler.NewLoginHandler(storage)

	router.Post("/api/user/register", registerHandler.Register)
	router.Post("/api/user/login", loginHandler.Login)

	return router, nil
}
