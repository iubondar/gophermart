package service

import "github.com/go-resty/resty/v2"

type PollingService struct {
	client *resty.Client
}

func NewPollingService(accrualSystemAddress string) *PollingService {
	client := resty.New().SetBaseURL(accrualSystemAddress)

	return &PollingService{
		client: client,
	}
}

func (s PollingService) Start() {

}
