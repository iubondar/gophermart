package config

import (
	"flag"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
)

type Config struct {
	RunAddress           string        `env:"RUN_ADDRESS"`
	DatabaseURI          string        `env:"DATABASE_URI"`
	AccrualSystemAddress string        `env:"ACCRUAL_SYSTEM_ADDRESS"`
	PollingInterval      time.Duration `env:"POLLING_INTERVAL"`
	FetchingLimit        int           `env:"FETCHING_LIMIT"`
}

// для локальной разработки
const (
	defaultRunAddress           = "localhost:8000"
	defaultDatabaseURI          = "host=localhost user=newuser password=password dbname=gophermart sslmode=disable"
	defaultAccrualSystemAddress = "http://localhost:8080"
	defaultPollingInterval      = 1 * time.Second
	defaultFetchingLimit        = 10
)

func NewConfig(progname string, args []string) (*Config, error) {
	var c Config

	// https://eli.thegreenplace.net/2020/testing-flag-parsing-in-go-programs/
	// Загружаем значения из переданных аргументов командной строки
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	flags.StringVar(&c.RunAddress, "a", defaultRunAddress, "address to run server")
	flags.StringVar(&c.DatabaseURI, "d", defaultDatabaseURI, "database dsn")
	flags.StringVar(&c.AccrualSystemAddress, "r", defaultAccrualSystemAddress, "address to connect to accrual service")
	flags.DurationVar(&c.PollingInterval, "i", defaultPollingInterval, "polling interval for accrual service")
	flags.IntVar(&c.FetchingLimit, "l", defaultFetchingLimit, "limit of orders to fetch in one request")

	err := flags.Parse(args)
	if err != nil {
		return nil, err
	}

	// Переписываем значения из переменных окружения
	err = env.Parse(&c)
	if err != nil {
		return nil, err
	}

	zap.L().Sugar().Debugln(
		"Config: ",
		"RunAddress", c.RunAddress,
		"DatabaseURI", c.DatabaseURI,
		"AccrualSystemAddress", c.AccrualSystemAddress,
		"PollingInterval", c.PollingInterval,
		"FetchingLimit", c.FetchingLimit,
	)

	return &c, nil
}
