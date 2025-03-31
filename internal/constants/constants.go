package constants

type OrderProcessingStatus string

const (
	New        OrderProcessingStatus = "NEW"
	Processing OrderProcessingStatus = "PROCESSIMG"
	Invalid    OrderProcessingStatus = "INVALID"
	Processed  OrderProcessingStatus = "PROCESSED"
)

type OrderRegistrationResult int

const (
	AlreadyRegistered OrderRegistrationResult = iota + 1
	AcceptedToProcessing
	RegisteredByAnotherUser
	WrongOrderNumberFormat
)
