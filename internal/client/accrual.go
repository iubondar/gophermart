package client

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/iubondar/gophermart/internal/constants"
	"go.uber.org/zap"
)

type (
	AccrualClient struct {
		httpc *resty.Client
	}

	OrderStatusOut struct {
		Number  string
		Status  constants.OrderStatus
		Accrual int
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

func (c AccrualClient) FetchOrderStatus(orderNumber string) (out OrderStatusOut, err error) {
	var result accrualStatus
	resp, err := c.httpc.R().SetResult(&result).Get("/api/orders/" + orderNumber)
	if err != nil {
		zap.L().Sugar().Debugln("Error fetching order status, number ", orderNumber, ", error: ", err.Error())
		return OrderStatusOut{}, err
	}

	if resp.StatusCode() >= 400 {
		return OrderStatusOut{}, fmt.Errorf("fetching order status: %s", resp.Status())
	}

	if resp.StatusCode() == http.StatusNoContent {
		return OrderStatusOut{
			Number:  orderNumber,
			Status:  constants.OrderStatusNew,
			Accrual: 0,
		}, nil
	}

	return OrderStatusOut{
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
