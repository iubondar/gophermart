package service

import (
	"context"
	"time"

	"github.com/iubondar/gophermart/internal/models"
	"go.uber.org/zap"
)

const defaultPollingInterval = 5 * time.Second
const fetchingLimit = 10

type OrderStatusFetcher interface {
	FetchOrderStatus(order models.Order) (out models.OrderStatus, err error)
}

type OrdersRepository interface {
	OrdersToUpdate(ctx context.Context, limit int) (orders []models.Order, err error)
	UpdateOrders(ctx context.Context, orders []models.OrderStatus) error
}
type PollingService struct {
	fetcher         OrderStatusFetcher
	repo            OrdersRepository
	pollingInterval time.Duration
	doneCh          chan struct{}
}

func NewPollingService(fetcher OrderStatusFetcher, repo OrdersRepository, pollingInterval time.Duration) *PollingService {
	return &PollingService{
		fetcher:         fetcher,
		repo:            repo,
		pollingInterval: pollingInterval,
		doneCh:          make(chan struct{}),
	}
}

func (s PollingService) Start() {
	if s.pollingInterval == 0 {
		s.pollingInterval = defaultPollingInterval
	}
	ticker := time.NewTicker(s.pollingInterval)

	go func() {
		for {
			select {
			case <-s.doneCh:
				ticker.Stop()
				return
			case <-ticker.C:
				s.updateNextOrders(fetchingLimit)
			}
		}
	}()
}

func (s PollingService) Stop() {
	close(s.doneCh)
}

func (s PollingService) updateNextOrders(limit int) {
	// извлекаем из репозитория номера заказов для обновления
	orders, err := s.repo.OrdersToUpdate(context.Background(), limit)
	if err != nil {
		zap.L().Sugar().Debugln("Error fetching orders to update status, error: ", err.Error())
		return
	}

	// запрашиваем в цикле статус из Accrual системы
	var orderStatuses []models.OrderStatus
	for _, order := range orders {
		orderStatus, err := s.fetcher.FetchOrderStatus(order)
		if err != nil {
			zap.L().Sugar().Debugln("Error fetching orders to update status, error: ", err.Error())
			return
		}
		orderStatuses = append(orderStatuses, orderStatus)
	}

	// обновляем данные по заказам и баланс пользователя
	err = s.repo.UpdateOrders(context.Background(), orderStatuses)
	if err != nil {
		zap.L().Sugar().Debugln("Error updating order statuses, error: ", err.Error())
		return
	}
}
