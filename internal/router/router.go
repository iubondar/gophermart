package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/gophermart/internal/handler"
	"github.com/iubondar/gophermart/internal/logging"
	"github.com/iubondar/gophermart/internal/storage"
)

func NewRouter(storage *storage.Storage) (chi.Router, error) {
	registerHandler := handler.NewRegisterHandler(storage)
	loginHandler := handler.NewLoginHandler(storage)
	registerOrderHandler := handler.NewRegisterOrderHandler(storage)
	ordersHandler := handler.NewOrdersHandler(storage)
	withrawalsHandler := handler.NewWithdrawalsHandler(storage)

	router := chi.NewRouter()
	router.Use(logging.WithLogging)
	router.Post("/api/user/register", registerHandler.Register)
	router.Post("/api/user/login", loginHandler.Login)
	router.Post("/api/user/orders", registerOrderHandler.RegisterOrder)
	router.Get("/api/user/orders", ordersHandler.Orders)
	router.Get("/api/user/withdrawals", withrawalsHandler.Withdrawals)

	return router, nil
}
