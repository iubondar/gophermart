package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
}

func New(addr string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Start запускает сервер и начинает обработку запросов
func (s *Server) Start() error {
	// Канал для обработки ошибок сервера
	serverErrors := make(chan error, 1)

	// Запускаем сервер в отдельной горутине
	go func() {
		zap.L().Info("Starting server", zap.String("address", s.httpServer.Addr))
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Канал для обработки сигналов завершения от ОС
	shutdown := make(chan os.Signal, 1)
	// Регистрируем обработчики для SIGINT (Ctrl+C) и SIGTERM
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Ожидаем либо ошибку сервера, либо сигнал завершения
	select {
	case err := <-serverErrors:
		zap.L().Error("Server error", zap.Error(err))
		return err

	case sig := <-shutdown:
		zap.L().Info("Start shutdown", zap.String("signal", sig.String()))
		return s.Shutdown()
	}
}

// Shutdown выполняет graceful shutdown сервера
func (s *Server) Shutdown() error {
	// Устанавливаем таймаут 5 секунд для завершения текущих запросов
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся корректно завершить работу сервера
	if err := s.httpServer.Shutdown(ctx); err != nil {
		zap.L().Error("Graceful shutdown did not complete", zap.Error(err))
		// Если плавное завершение не удалось, принудительно закрываем сервер
		if err := s.httpServer.Close(); err != nil {
			zap.L().Error("Could not stop server", zap.Error(err))
			return err
		}
		return err
	}

	return nil
}
