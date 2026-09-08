package prometheus

import (
	"net/http"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

type PrometheusResponse struct {
	Status    string `json:"status"`
	IsPartial bool   `json:"isPartial"`
	Data      struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Host string `json:"host"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// PrometheusClient представляет клиент для работы
type PrometheusClient struct {
	baseURL    string
	auth       string
	client     *http.Client
	logs       *logger.LogWriter
	debug      bool
	start      time.Time
	end        time.Time
	Timeperiod string
}
