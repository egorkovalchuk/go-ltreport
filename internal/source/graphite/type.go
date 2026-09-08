package graphite

import (
	"net/http"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// Datapoint представляет одну точку данных Graphite
type Datapoint struct {
	Value     float64
	Timestamp int64
}

// MetricResponse представляет ответ Graphite API
type MetricResponse struct {
	Target     string      `json:"target"`
	Datapoints []Datapoint `json:"datapoints"`
}

// GraphiteClient представляет клиент для работы с Graphite API
type GraphiteClient struct {
	baseURL string
	client  *http.Client
	auth    string
	logs    *logger.LogWriter
	debug   bool
}
