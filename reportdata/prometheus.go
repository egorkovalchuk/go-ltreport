package reportdata

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
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
	baseURL string
	auth    string
	client  *http.Client
	logFunc func(string, interface{})
	debug   bool
}

// NewPrometheusClient создает новый экземпляр клиента
func NewPrometheusClient(baseURL string, auth string, logFunc func(string, interface{}), debug bool) *PrometheusClient {
	return &PrometheusClient{
		baseURL: baseURL,
		auth:    auth,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (p *PrometheusClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *PrometheusClient) ProcessDebug(t interface{}) {
	if p.debug {
		p.logFunc("DEBUG", t)
	}
}

func (p *PrometheusClient) GetDataMean(query string) (PrometheusResponse, error) {
	req, err := http.NewRequest("GET", p.baseURL+"/api/v1/query?query="+query, nil)
	if err != nil {
		return PrometheusResponse{}, fmt.Errorf("GetDataMean request failed: %v", err)
	}
	req.Header.Add("Authorization", p.auth)
	resp, err := p.client.Do(req)

	if err != nil {
		return PrometheusResponse{}, fmt.Errorf("GetDataMean request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		p.logFunc("INFO", "Request prometheus threshold success")
		var prom PrometheusResponse
		err = prom.JsonPrometheusParse(resp)
		if err != nil {
			return PrometheusResponse{}, fmt.Errorf("PROMETEUS: Error parse: %v", err)
		} else {
			return prom, nil
		}
	} else {
		return PrometheusResponse{}, fmt.Errorf("PROMETEUS: Request prometheus threshold error " + strconv.Itoa(resp.StatusCode) + " " + p.baseURL)
	}
}

func (p *PrometheusClient) GetThreshold(query string) (float64, error) {
	var percentile float64
	p.ProcessDebug("Get Prometheus threshold request: " + p.baseURL + "/api/v1/query?query=" + query)
	prom, err := p.GetDataMean(query)
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
		return fmt.Errorf("PROMETEUS: %v", err)
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

	if field != nil {
		if ff, ok := field.(string); ok {
			fieldp.ValString = ff
			if s, err := strconv.ParseFloat(ff, 64); err == nil {
				fieldp.ValFloat = s
			} else {
				fieldp.ValFloat = 0
			}
			if s, err := strconv.ParseInt(ff, 0, 64); err == nil {
				fieldp.ValInt = s
			} else {
				fieldp.ValInt = 0
			}
		} else {
			fieldp.ValFloat = 0
		}

	} else {
		fieldp.ValFloat = 0
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
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValString
}
