package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load(t *testing.T) {

	tests := []struct {
		name    string
		args    []string
		envVars Config
		want    Config
	}{
		{
			name:    "Defaults",
			args:    nil,
			envVars: Config{RunAddress: "", DatabaseURI: "", AccrualSystemAddress: "", PollingInterval: 0, FetchingLimit: 0},
			want: Config{
				RunAddress:           defaultRunAddress,
				DatabaseURI:          defaultDatabaseURI,
				AccrualSystemAddress: defaultAccrualSystemAddress,
				PollingInterval:      defaultPollingInterval,
				FetchingLimit:        defaultFetchingLimit,
			},
		},
		{
			name:    "Override with flags",
			args:    []string{"-a", "localhost:8888", "-d", "host=local user=u password=p dbname=db", "-r", "localhost:8800", "-i", "2s", "-l", "20"},
			envVars: Config{RunAddress: "", DatabaseURI: "", AccrualSystemAddress: "", PollingInterval: 0, FetchingLimit: 0},
			want: Config{
				RunAddress:           "localhost:8888",
				DatabaseURI:          "host=local user=u password=p dbname=db",
				AccrualSystemAddress: "localhost:8800",
				PollingInterval:      2 * time.Second,
				FetchingLimit:        20,
			},
		},
		{
			name: "Override with envs",
			args: []string{"-a", "localhost:8888", "-d", "host=local user=u password=p dbname=db", "-r", "localhost:8800", "-i", "2s", "-l", "20"},
			envVars: Config{
				RunAddress:           "localhost:8800",
				DatabaseURI:          "host=local user=uu password=pp dbname=db123",
				AccrualSystemAddress: "localhost:8888",
				PollingInterval:      3 * time.Second,
				FetchingLimit:        30,
			},
			want: Config{
				RunAddress:           "localhost:8800",
				DatabaseURI:          "host=local user=uu password=pp dbname=db123",
				AccrualSystemAddress: "localhost:8888",
				PollingInterval:      3 * time.Second,
				FetchingLimit:        30,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("RUN_ADDRESS", tt.envVars.RunAddress)
			t.Setenv("DATABASE_URI", tt.envVars.DatabaseURI)
			t.Setenv("ACCRUAL_SYSTEM_ADDRESS", tt.envVars.AccrualSystemAddress)
			if tt.envVars.PollingInterval > 0 {
				t.Setenv("POLLING_INTERVAL", tt.envVars.PollingInterval.String())
			} else {
				t.Setenv("POLLING_INTERVAL", "")
			}
			if tt.envVars.FetchingLimit > 0 {
				t.Setenv("FETCHING_LIMIT", strconv.Itoa(tt.envVars.FetchingLimit))
			} else {
				t.Setenv("FETCHING_LIMIT", "")
			}

			c, err := NewConfig("Test", tt.args)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, *c)
		})
	}
}
