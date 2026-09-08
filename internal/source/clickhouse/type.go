package clickhouse

import (
	"net/http"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// Result of Query.
type Result interface {
	//	DecodeResult(r *Reader, version int, b Block) error
}

type MetaStruct struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Len  int
}
type ClickHouseJson struct {
	ColunmLen int
	Name      string
	Meta      []MetaStruct `json:"meta"`
	Data      []map[string]interface {
	} `json:"data"`
	Rows       int `json:"rows"`
	Statistics struct {
		Elapsed    float32 `json:"elapsed"`
		Rows_read  float32 `json:"rows_read"`
		Bytes_read float32 `json:"bytes_read"`
	} `json:"statistics"`
}

// CHClient представляет клиент для работы с Grafana
type CHClient struct {
	baseURL    string
	user       string
	pass       string
	client     *http.Client
	logs       *logger.LogWriter
	debug      bool
	start      time.Time
	end        time.Time
	Timeperiod string
}
