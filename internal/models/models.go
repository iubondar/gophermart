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
	Accrual    float32
	UploadedAt time.Time
}

type OrderStatus struct {
	UserID  uuid.UUID
	Number  string
	Status  constants.OrderStatus
	Accrual float32
}

type Withdrawal struct {
	Number      string
	Sum         float32
	ProcessedAt time.Time
}

type Account struct {
	Balance       float32
	WithdrawalSum float32
}
