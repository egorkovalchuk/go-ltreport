package reportdata

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
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

// NewPrometheusClient создает новый экземпляр клиента
func NewPrometheusClient(baseURL string, auth string, start, end time.Time, logs *logger.LogWriter, debug bool) *PrometheusClient {
	return &PrometheusClient{
		baseURL:    baseURL,
		auth:       auth,
		client:     &http.Client{Timeout: 30 * time.Second},
		logs:       logs,
		debug:      debug,
		start:      start,
		end:        end,
		Timeperiod: `&start=` + fmt.Sprintf("%d", start.Unix()) + `&end=` + fmt.Sprintf("%d", end.Unix()),
	}
}

func (p *PrometheusClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *PrometheusClient) GetDataMean(query string) (PrometheusResponse, error) {
	req, err := http.NewRequest("GET", p.baseURL+"/api/v1/query?query="+query, nil)
	if err != nil {
		return PrometheusResponse{}, fmt.Errorf("GetDataMean request failed: %w", err)
	}
	req.Header.Add("Authorization", p.auth)
	resp, err := p.client.Do(req)

	if err != nil {
		return PrometheusResponse{}, fmt.Errorf("GetDataMean request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		p.logs.ProcessInfo("Request prometheus threshold success")
		var prom PrometheusResponse
		err = prom.JsonPrometheusParse(resp)
		if err != nil {
			return PrometheusResponse{}, fmt.Errorf("PROMETEUS: Error parse: %w", err)
		} else {
			return prom, nil
		}
	} else {
		return PrometheusResponse{}, fmt.Errorf("PROMETEUS: Request prometheus threshold error %s %s", strconv.Itoa(resp.StatusCode), p.baseURL)
	}
}

func (p *PrometheusClient) GetThreshold(query string) (float64, error) {
	var percentile float64
	p.logs.ProcessDebug("Get Prometheus threshold request: " + p.baseURL + "/api/v1/query?query=" + query + p.Timeperiod)
	prom, err := p.GetDataMean(query + p.Timeperiod)
	if err == nil {
		percentile = prom.JsonPrometheusFiledParseFloat(prom.Data.Result[0].Value[1])
		return percentile, nil
	} else {
		return 0, err
	}
}

func (p *PrometheusResponse) JsonPrometheusParse(resp *http.Response) error {

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&p)

	if err != nil {
		return fmt.Errorf("PROMETEUS: %w", err)
	}

	if p.Status != "success" {
		return fmt.Errorf("PROMETEUS: Expected exactly one result in response, got %s", p.Status)
	}

	if len(p.Data.Result) == 0 {
		return fmt.Errorf("PROMETEUS: Expected exactly one series in result, got %d", len(p.Data.Result))
	}

	return nil

}

func (p *PrometheusResponse) JsonPrometheusFiledParse(field interface{}) SField {
	var fieldp SField

	if field == nil {
		return fieldp // все поля SField уже инициализированы нулевыми значениями
	}

	switch v := field.(type) {
	case string:
		fieldp.ValString = v
		// Пытаемся преобразовать строку в числа
		if s, err := strconv.ParseFloat(v, 64); err == nil {
			fieldp.ValFloat = s
		}
		if s, err := strconv.ParseInt(v, 0, 64); err == nil {
			fieldp.ValInt = s
		}

	case float64:
		fieldp.ValFloat = v
		fieldp.ValInt = int64(v)

	case float32:
		fieldp.ValFloat = float64(v)
		fieldp.ValInt = int64(v)

	case int:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case int64:
		fieldp.ValInt = v
		fieldp.ValFloat = float64(v)

	case int32:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case int16:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case int8:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case uint:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case uint64:
		// Осторожно с большими uint64 значениями при конвертации в int64
		if v <= math.MaxInt64 {
			fieldp.ValInt = int64(v)
		}
		fieldp.ValFloat = float64(v)

	case uint32:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case uint16:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case uint8:
		fieldp.ValInt = int64(v)
		fieldp.ValFloat = float64(v)

	case bool:
		fieldp.ValString = strconv.FormatBool(v)
		if v {
			fieldp.ValInt = 1
			fieldp.ValFloat = 1.0
		}

	case json.Number:
		fieldp.ValString = v.String()
		if s, err := strconv.ParseFloat(fieldp.ValString, 64); err == nil {
			fieldp.ValFloat = s
		}
		if s, err := strconv.ParseInt(fieldp.ValString, 0, 64); err == nil {
			fieldp.ValInt = s
		}
	}

	return fieldp
}

func (p *PrometheusResponse) JsonPrometheusFiledParseFloat(field interface{}) float64 {
	fieldp := p.JsonPrometheusFiledParse(field)
	return fieldp.ValFloat
}

func (p *PrometheusResponse) JsonPrometheusFiledParseInt(field interface{}) int64 {
	fieldp := p.JsonPrometheusFiledParse(field)
	return fieldp.ValInt
}

func (p *PrometheusResponse) JsonPrometheusFiledParseString(field interface{}) string {
	fieldp := p.JsonPrometheusFiledParse(field)
	return fieldp.ValString
}
