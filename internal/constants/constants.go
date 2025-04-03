package constants

type OrderStatus string // статусы в нашей базе

const (
	OrderStatusNew        OrderStatus = "NEW"        // заказ зарегистрирован, либо ещё не проверяли его статус в accrual сервисе, либо там статус REGISTERED
	OrderStatusProcessing OrderStatus = "PROCESSING" // статус от момента получения PROCESSIMG из accrual сервиса до перехода в терминальный статус
	OrderStatusInvalid    OrderStatus = "INVALID"    // терминальный статус - начисление не будет произведено
	OrderStatusProcessed  OrderStatus = "PROCESSED"  // терминальный статус - получили баллы от accrual сервиса и произвели начисление
)

type OrderRegistrationResult int

const (
	AlreadyRegistered OrderRegistrationResult = iota + 1
	AcceptedToProcessing
	RegisteredByAnotherUser
	WrongOrderNumberFormat
)

type WithdrawResult int

const (
	Success WithdrawResult = iota + 1
	InsufficientFunds
	WrongOrderFormat
)
