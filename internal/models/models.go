package models

import (
	"time"

	"github.com/iubondar/gophermart/internal/constants"
)

type Order struct {
	Number     string
	Status     constants.OrderProcessingStatus
	Accrual    int
	UploadedAt time.Time
}
