package service

import (
	"context"
	"sync"
	"time"

	"github.com/iubondar/gophermart/internal/models"
	"go.uber.org/zap"
)

type OrderStatusFetcher interface {
	FetchOrderStatus(ctx context.Context, order models.Order) (out models.OrderStatus, err error)
}

type OrdersRepository interface {
	OrdersToUpdate(ctx context.Context, limit int) (orders []models.Order, err error)
	UpdateOrders(ctx context.Context, orders []models.OrderStatus) error
}

type Result struct {
	Status models.OrderStatus
	Err    error
}

type PollingService struct {
	interval    time.Duration
	concurrency int
	fetcher     OrderStatusFetcher
	repo        OrdersRepository
	done        chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewPollingService(
	interval time.Duration,
	concurrency int,
	fetcher OrderStatusFetcher,
	repo OrdersRepository,
) *PollingService {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingService{
		interval:    interval,
		concurrency: concurrency,
		fetcher:     fetcher,
		repo:        repo,
		done:        make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (ps *PollingService) Start() {
	ticker := time.NewTicker(ps.interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				ps.runPollingCycle()
			case <-ps.done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (ps *PollingService) Stop() {
	close(ps.done)
	ps.cancel()
	ps.wg.Wait()
}

func (ps *PollingService) runPollingCycle() {
	// извлекаем из репозитория номера заказов для обновления
	orders, err := ps.repo.OrdersToUpdate(ps.ctx, ps.concurrency)
	if err != nil {
		zap.L().Sugar().Debugln("Error fetching orders to update status, error: ", err.Error())
		return
	}

	// Создаём input и output каналы
	ordersCh := make(chan models.Order, len(orders))
	resultCh := make(chan Result, len(orders))

	// Заполняем input канал заказами для обновления статуса
	go func() {
		defer close(ordersCh)
		for _, order := range orders {
			ordersCh <- order
		}
	}()

	// Fan-out: запускаем воркеры
	ps.wg.Add(ps.concurrency)
	for range ps.concurrency {
		go ps.worker(ordersCh, resultCh)
	}

	// Fan-in: собираем все результаты
	go func() {
		ps.wg.Wait()    // ждём завершения всех воркеров
		close(resultCh) // закрываем канал

		// Обрабатываем все результаты после завершения
		var orderStatuses []models.OrderStatus
		for result := range resultCh {
			if result.Err != nil {
				zap.L().Sugar().Debugln("Got error from worker: ", result.Err.Error())
				// Even if there's an error, we still want to include the status in the update
				orderStatuses = append(orderStatuses, result.Status)
			} else {
				orderStatuses = append(orderStatuses, result.Status)
			}
		}

		// Always call UpdateOrders, even if there are no statuses to update
		err = ps.repo.UpdateOrders(ps.ctx, orderStatuses)
		if err != nil {
			zap.L().Sugar().Debugln("Error updating order statuses, error: ", err.Error())
			return
		}
	}()
}

func (ps *PollingService) worker(ordersCh <-chan models.Order, resultsCh chan<- Result) {
	defer ps.wg.Done()

	for order := range ordersCh {
		status, err := ps.fetcher.FetchOrderStatus(ps.ctx, order)
		resultsCh <- Result{Status: status, Err: err}
	}
}
