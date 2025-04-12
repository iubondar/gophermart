package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	pollingService := service.NewPollingService(0, 0, accrualClient, storage)
	pollingService.Start()

	router, err := router.NewRouter(storage)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем HTTP сервер с указанным адресом и роутером
	server := &http.Server{
		Addr:    config.RunAddress,
		Handler: router,
	}

	// Канал для обработки ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		zap.L().Info("Starting server", zap.String("address", config.RunAddress))
		serverErrors <- server.ListenAndServe()
	}()

	// Канал для обработки сигналов завершения от ОС
	shutdown := make(chan os.Signal, 1)
	// Регистрируем обработчики для SIGINT (Ctrl+C) и SIGTERM
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Ожидаем либо ошибку сервера, либо сигнал завершения
	select {
	case err := <-serverErrors:
		zap.L().Error("Server error", zap.Error(err))
		os.Exit(1)

	case sig := <-shutdown:
		zap.L().Info("Start shutdown", zap.String("signal", sig.String()))

		// Устанавливаем таймаут 5 секунд для завершения текущих запросов
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Останавливаем сервис опроса
		pollingService.Stop()

		// Пытаемся корректно завершить работу сервера
		if err := server.Shutdown(ctx); err != nil {
			zap.L().Error("Graceful shutdown did not complete", zap.Error(err))
			// Если плавное завершение не удалось, принудительно закрываем сервер
			if err := server.Close(); err != nil {
				zap.L().Error("Could not stop server", zap.Error(err))
			}
		}
	}
}
