package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/iubondar/gophermart/internal/constants"
)

type Order struct {
	UserID     uuid.UUID
	Number     string
	Status     constants.OrderStatus
	Accrual    int
	UploadedAt time.Time
}

type OrderStatus struct {
	UserID  uuid.UUID
	Number  string
	Status  constants.OrderStatus
	Accrual int
}

type Withdrawal struct {
	Number      string
	Sum         int
	ProcessedAt time.Time
}
