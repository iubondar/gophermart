package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iubondar/gophermart/internal/client"
	"github.com/iubondar/gophermart/internal/config"
	"github.com/iubondar/gophermart/internal/router"
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

	pollingService := service.NewPollingService(accrualClient, storage, 0)
	pollingService.Start()

	router, err := router.NewRouter(storage)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(
		http.ListenAndServe(config.RunAddress, router),
	)
}
