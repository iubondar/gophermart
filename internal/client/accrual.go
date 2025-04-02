package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/iubondar/gophermart/internal/constants"
	"github.com/iubondar/gophermart/internal/models"
	"go.uber.org/zap"
)

type (
	AccrualClient struct {
		httpc *resty.Client
	}

	accrualStatus struct {
		Order   string `json:"order"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual"`
	}
)

func NewAccrualClient(accrualSystemAddress string) *AccrualClient {
	client := resty.New().SetBaseURL(accrualSystemAddress)

	return &AccrualClient{
		httpc: client,
	}
}

func (c AccrualClient) FetchOrderStatus(order models.Order) (out models.OrderStatus, err error) {
	var result accrualStatus
	resp, err := c.httpc.R().SetResult(&result).Get("/api/orders/" + order.Number)
	if err != nil {
		zap.L().Sugar().Debugln("Error fetching order status, number ", order.Number, ", error: ", err.Error())
		return models.OrderStatus{}, err
	}

	if resp.StatusCode() >= 400 {
		return models.OrderStatus{}, fmt.Errorf("fetching order status: %s", resp.Status())
	}

	if resp.StatusCode() == http.StatusNoContent {
		return models.OrderStatus{
			UserID:  order.UserID,
			Number:  result.Order,
			Status:  constants.OrderStatusNew,
			Accrual: 0,
		}, nil
	}

	return models.OrderStatus{
		UserID:  order.UserID,
		Number:  result.Order,
		Status:  mapAccrualStatus(result.Status),
		Accrual: result.Accrual,
	}, nil
}

func mapAccrualStatus(accrualStatus string) constants.OrderStatus {
	switch accrualStatus {
	case "REGISTERED":
		return constants.OrderStatusProcessing
	case "INVALID":
		return constants.OrderStatusInvalid
	case "PROCESSING":
		return constants.OrderStatusProcessing
	case "PROCESSED":
		return constants.OrderStatusProcessed
	default:
		return constants.OrderStatusNew
	}
}

func (c AccrualClient) UpdateOrders(ctx context.Context, orders []models.OrderStatus) {

}
