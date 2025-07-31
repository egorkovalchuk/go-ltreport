package reportdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"
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
	logFunc func(string, interface{})
	debug   bool
}

// NewGraphiteClient создает новый экземпляр клиента
func NewGraphiteClient(baseURL string, auth string, logFunc func(string, interface{}), debug bool) *GraphiteClient {
	return &GraphiteClient{
		baseURL: baseURL,
		auth:    auth,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (gc *GraphiteClient) Close() {
	gc.client.CloseIdleConnections()
}

func (gc *GraphiteClient) ProcessDebug(t interface{}) {
	if gc.debug {
		gc.logFunc("DEBUG", t)
	}
}

// GetMetrics получает метрики из Graphite API
func (gc *GraphiteClient) GetMetrics(target string, from, until time.Time) ([]MetricResponse, error) {

	requestURL := gc.baseURL + "/render?" + "target=" + target + "&from=" + fmt.Sprintf("%d", from.Unix()) + "&until=" + fmt.Sprintf("%d", until.Unix()) + "&format=json"
	//gc.ProcessDebug(requestURL)
	rsp, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		gc.ProcessDebug("Get Graphite threshold request: " + requestURL)
		return nil, fmt.Errorf("request failed: %v", err)
	}
	rsp.Header.Add("Authorization", gc.auth)

	resp, err := gc.client.Do(rsp)

	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		gc.ProcessDebug("Get Graphite threshold request: " + requestURL)
		return nil, fmt.Errorf("graphite API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Временная структура для разбора исходного формата Graphite
	var rawResponse []struct {
		Target     string          `json:"target"`
		Datapoints [][]interface{} `json:"datapoints"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Преобразуем в наш формат
	var result []MetricResponse
	for _, item := range rawResponse {
		var dps []Datapoint
		for _, dp := range item.Datapoints {
			if len(dp) != 2 {
				continue
			}

			value, ok1 := dp[0].(float64)
			timestamp, ok2 := dp[1].(float64)
			if !ok1 || !ok2 {
				continue
			}

			dps = append(dps, Datapoint{
				Value:     value,
				Timestamp: int64(timestamp),
			})
		}

		result = append(result, MetricResponse{
			Target:     item.Target,
			Datapoints: dps,
		})
	}

	return result, nil
}

// ListMetrics получает список доступных метрик
func (gc *GraphiteClient) ListMetrics() ([]string, error) {
	requestURL := fmt.Sprintf("%s/metrics", gc.baseURL)

	resp, err := gc.client.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("graphite API returned status %d: %s", resp.StatusCode, string(body))
	}
	var metrics []string
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return metrics, nil
}

// Get99thPercentile вычисляет 99-й персентиль для указанной метрики
func (gc *GraphiteClient) Get99thPercentile(target string, from, until time.Time) (float64, error) {
	metrics, err := gc.GetMetrics(target, from, until)
	if err != nil {
		return 0, err
	}

	if len(metrics) == 0 {
		return 0, fmt.Errorf("no metrics found for target %s", target)
	}

	// Собираем все значения
	var values []float64
	for _, dp := range metrics[0].Datapoints {
		if !math.IsNaN(dp.Value) {
			values = append(values, dp.Value)
		}
	}

	if len(values) == 0 {
		return 0, fmt.Errorf("no valid data points found")
	}

	// Вычисляем 99-й персентиль
	return CalculatePercentile(values, 99), nil
}
