package reportdata

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

var (
	// режим Дебага
	DebugFlag = false
	LogFlag   = false
	err       error
)

//структура influx type 1
type Mean struct {
	Results []struct {
		StatementID int `json:"statement_id"`
		Series      []struct {
			Name string `json:"name"`
			Tags struct {
				Transaction string `json:"transaction"`
				Suite       string `json:"suite"`
				Statut      string `json:"statut"`
				Application string `json:"application"`
			} `json:"tags"`
			Columns []string        `json:"columns"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}

//структура influx type 2 для сценария
type MeanScenario struct {
	Results []struct {
		StatementID int `json:"statement_id"`
		Series      []struct {
			Name string `json:"name"`
			Tags struct {
				Transaction string `json:"transaction"`
				Suite       string `json:"suite"`
				Statut      string `json:"statut"`
				Application string `json:"application"`
			} `json:"tags"`
			Columns []string        `json:"columns"`
			Values  [][]interface{} `json:"values"`
		} `json:"series"`
	} `json:"results"`
}

//для преобразования типа ответа инфлюкса
type SField struct {
	NameCol   string
	ValFloat  float64
	ValInt    int64
	ValString string
	ValTime   int64
}

//Структруа ответа инфлюкса по тестам
type LTTestDinamic struct {
	NameTest string
	Field    []YField
}

//для хранения поля ответа
type YField struct {
	Name        string
	Value       float64
	Description string
	Statut      string
}

// InfluxClient представляет клиент для работы
type InfluxClient struct {
	baseURL string
	auth    string
	client  *http.Client
	logFunc func(string, interface{})
	debug   bool
}

// NewPrometheusClient создает новый экземпляр клиента
func NewInfluxClient(baseURL string, auth string, logFunc func(string, interface{}), debug bool) *InfluxClient {
	return &InfluxClient{
		baseURL: baseURL,
		auth:    auth,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (p *InfluxClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *InfluxClient) ProcessDebug(t interface{}) {
	if p.debug {
		p.logFunc("DEBUG", t)
	}
}

func (p *InfluxClient) GetThreshold(query string) (float64, error) {
	var percentile float64
	metrics, err := p.GetDataMean(query)
	if err == nil {
		percentile = JsonINfluxFiledParseFloat(metrics.Results[0].Series[0].Values[0][1])
		return percentile, nil
	} else {
		return 0, err
	}
}

func (p *InfluxClient) GetDataMean(query string) (Mean, error) {
	resp_inf, err := http.NewRequest("GET", p.baseURL+""+query, nil)
	if err != nil {
		return Mean{}, fmt.Errorf("GetDataMean request failed: %v", err)
	}
	if p.auth != "" {
		resp_inf.Header.Add("Authorization", p.auth)
	}
	p.logFunc("INFO", "Influx request "+p.baseURL+""+query)

	rsp_inf, err := p.client.Do(resp_inf)
	if err != nil {
		return Mean{}, fmt.Errorf("GetDataMean request failed: %v", err)
	}
	defer rsp_inf.Body.Close()

	if rsp_inf.StatusCode == http.StatusOK {
		p.logFunc("INFO", "Request Influx success")
		infjson, err := JsonINfluxParse(rsp_inf)
		if err == nil {
			return infjson, nil
		} else {
			return Mean{}, fmt.Errorf("GetDataMean error parse: %v", err)
		}
	} else {
		return Mean{}, fmt.Errorf("influx API returned status %d: %s", rsp_inf.StatusCode, p.baseURL+query)
	}
}

// Get99thPercentile вычисляет 99-й персентиль для указанной метрики
func (p *InfluxClient) Get99thPercentile(query string) (float64, error) {
	metrics, err := p.GetDataMean(query)
	if err != nil {
		return 0, err
	}

	if len(metrics.Results) == 0 || len(metrics.Results[0].Series) == 0 {
		return 0, fmt.Errorf("no metrics found for query %s", query)
	}

	// Собираем все значения
	var values []float64
	for _, dp := range metrics.Results[0].Series {
		for _, v := range dp.Values {
			value, ok1 := v[1].(float64)
			if !ok1 {
				continue
			}
			if !math.IsNaN(value) {
				values = append(values, value)
			}
		}
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("no valid data points found")
	}

	// Вычисляем 99-й персентиль
	return CalculatePercentile(values, 99), nil
}

func JsonINfluxFiledParse(field interface{}) SField {
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

func JsonINfluxFiledParseFloat(field interface{}) float64 {
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValFloat
}

func JsonINfluxFiledParseInt(field interface{}) int64 {
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValInt
}

func JsonINfluxParse(resp *http.Response) (Mean, error) {
	var infjson Mean

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&infjson)

	if err != nil {
		return infjson, fmt.Errorf("INFLUX: %v", err)
	}

	if len(infjson.Results) == 0 || len(infjson.Results[0].Series) == 0 {
		return infjson, fmt.Errorf("Expected exactly one result in response, got %d", len(infjson.Results))
	}

	return infjson, nil
}

func InfluxJmeterScenarioStatut(i [][]interface{}, statut string, cfg []JmeterQScnrFieldS) []YField {

	var LTTest_yfield []YField
	var LTTest_yfieldtmp YField

	num := 1
	for _, j := range cfg {

		LTTest_yfieldtmp.Name = j.Name
		LTTest_yfieldtmp.Description = j.Description
		LTTest_yfieldtmp.Statut = statut
		LTTest_yfieldtmp.Value = JsonINfluxFiledParseFloat(i[0][num])
		num++
		LTTest_yfield = append(LTTest_yfield, LTTest_yfieldtmp)
	}

	return LTTest_yfield

}
