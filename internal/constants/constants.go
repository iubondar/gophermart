package constants

type OrderProcessingStatus string

const (
	New        OrderProcessingStatus = "NEW"        // заказ зарегистрирован, либо ещё не проверяли его статус в accrual сервисе, либо там статус REGISTERED
	Processing OrderProcessingStatus = "PROCESSIMG" // статус от момента получения PROCESSIMG из accrual сервиса до перехода в терминальный статус
	Invalid    OrderProcessingStatus = "INVALID"    // терминальный статус - начисление не будет произведено
	Processed  OrderProcessingStatus = "PROCESSED"  // терминальный статус - получили баллы от accrual сервиса и произвели начисление
)

type OrderRegistrationResult int

const (
	AlreadyRegistered OrderRegistrationResult = iota + 1
	AcceptedToProcessing
	RegisteredByAnotherUser
	WrongOrderNumberFormat
)
