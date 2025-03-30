package constants

const (
	OrderStatusProcessing = "processing"
)

type OrderRegistrationResult int

const (
	AlreadyRegistered OrderRegistrationResult = iota + 1
	AcceptedToProcessing
	RegisteredByAnotherUser
	WrongOrderNumberFormat
)
