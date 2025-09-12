package reportdata

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// TelegrafClient представляет клиент для работы
type TelegrafClient struct {
	baseURL string
	auth    string
	client  *http.Client
	logFunc *logger.LogWriter
	debug   bool
}

// NewTelegrafClient создает новый экземпляр клиента
func NewTelegrafClient(baseURL string, auth string, logFunc *logger.LogWriter, debug bool) *TelegrafClient {
	return &TelegrafClient{
		baseURL: baseURL,
		auth:    auth,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (p *TelegrafClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *TelegrafClient) GetDataSourceThreshold(query string) (float64, error) {
	var percentile float64
	resp_inf, err := http.NewRequest("GET", p.baseURL+""+query, nil)
	if err != nil {
		return 0, err
	}
	resp_inf.Header.Add("Authorization", p.auth)

	p.logFunc.ProcessDebug("Get telegraf threshold request: " + p.baseURL + "/api/v1/query?query=" + query)
	rsp_inf, err := p.client.Do(resp_inf)

	if err != nil {
		return 0, err
	}
	defer rsp_inf.Body.Close()

	if rsp_inf.StatusCode == http.StatusOK {
		p.logFunc.ProcessInfo("Request prometheus threshold success")

		var prom PrometheusResponse
		err = prom.JsonPrometheusParse(rsp_inf)
		if err == nil {
			percentile = prom.JsonPrometheusFiledParseFloat(prom.Data.Result[0].Value[1])
		} else {
			return 0, err
		}

	} else {
		return 0, fmt.Errorf("Request prometheus threshold error " + strconv.Itoa(rsp_inf.StatusCode) + " " + p.baseURL)
	}

	return percentile, nil
}
