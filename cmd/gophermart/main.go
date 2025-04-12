package main

import (
	"log"
	"os"

	"github.com/iubondar/gophermart/internal/client"
	"github.com/iubondar/gophermart/internal/config"
	"github.com/iubondar/gophermart/internal/router"
	"github.com/iubondar/gophermart/internal/server"
	"github.com/iubondar/gophermart/internal/service"
	"github.com/iubondar/gophermart/internal/storage"
	"go.uber.org/zap"
)

func init() {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))
}

func main() {
	config, err := config.NewConfig(os.Args[0], os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	storage, err := storage.NewStorage(config.DatabaseURI)
	if err != nil {
		log.Fatal(err)
	}

	accrualClient := client.NewAccrualClient(config.AccrualSystemAddress)

	pollingService := service.NewPollingService(
		config.PollingInterval,
		config.FetchingLimit,
		accrualClient,
		storage,
	)
	pollingService.Start()

	router, err := router.NewRouter(storage)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем и запускаем сервер
	srv := server.New(config.RunAddress, router)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
