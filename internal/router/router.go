package router

import (
	"github.com/go-chi/chi"
	"github.com/iubondar/gophermart/internal/compress"
	"github.com/iubondar/gophermart/internal/handler"
	"github.com/iubondar/gophermart/internal/logging"
	"github.com/iubondar/gophermart/internal/storage"
	"github.com/iubondar/gophermart/internal/usecase"
)

func NewRouter(storage *storage.Storage) (chi.Router, error) {
	registerHandler := handler.NewRegisterHandler(storage)
	loginHandler := handler.NewLoginHandler(storage)
	registerOrderHandler := handler.NewRegisterOrderHandler(storage)
	ordersHandler := handler.NewOrdersHandler(storage)
	withdrawalsHandler := handler.NewWithdrawalsHandler(storage)
	withdrawHandler := handler.NewWithdrawHandler(storage)

	getBalanceUsecase := usecase.NewGetBalanceUsecase(storage)
	balanceHandler := handler.NewBalanceHandler(getBalanceUsecase)

	router := chi.NewRouter()
	router.Use(logging.WithLogging, compress.WithGzipCompression)

	router.Post("/api/user/register", registerHandler.Register)
	router.Post("/api/user/login", loginHandler.Login)

	router.Route("/api/user", func(r chi.Router) {
		r.Post("/orders", registerOrderHandler.RegisterOrder)
		r.Get("/orders", ordersHandler.Orders)

		r.Get("/balance", balanceHandler.Balance)
		r.Post("/balance/withdraw", withdrawHandler.Withdraw)

		r.Get("/withdrawals", withdrawalsHandler.Withdrawals)
	})

	return router, nil
}
