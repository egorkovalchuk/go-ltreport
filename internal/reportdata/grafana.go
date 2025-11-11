package reportdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// GrafanaClient представляет клиент для работы с Grafana
type GrafanaClient struct {
	baseURL    string
	auth       string
	client     *http.Client
	logs       *logger.LogWriter
	debug      bool
	start      time.Time
	end        time.Time
	Timeperiod string
}

// Структура алерта
type Alert struct {
	ID             int       `json:"Id"`
	Version        int       `json:"Version"`
	OrgID          int       `json:"OrgId"`
	DashboardID    int       `json:"DashboardId"`
	PanelID        int       `json:"PanelId"`
	Name           string    `json:"Name"`
	Message        string    `json:"Message"`
	Severity       string    `json:"Severity"`
	State          string    `json:"State"`
	Handler        int       `json:"Handler"`
	Silenced       bool      `json:"Silenced"`
	ExecutionError string    `json:"ExecutionError"`
	Frequency      int       `json:"Frequency"`
	For            int       `json:"For"`
	EvalData       EvalData  `json:"EvalData"`
	NewStateDate   time.Time `json:"NewStateDate"`
	StateChanges   int       `json:"StateChanges"`
	Created        time.Time `json:"Created"`
	Updated        time.Time `json:"Updated"`
	Settings       struct {
		AlertRuleTags struct {
		} `json:"alertRuleTags"`
		Conditions []struct {
			Evaluator struct {
				Params []int  `json:"params"`
				Type   string `json:"type"`
			} `json:"evaluator"`
			Operator struct {
				Type string `json:"type"`
			} `json:"operator"`
			Query struct {
				DatasourceID int `json:"datasourceId"`
				Model        struct {
					Alias      string `json:"alias"`
					Datasource struct {
						Type string `json:"type"`
						UID  string `json:"uid"`
					} `json:"datasource"`
					DsType string `json:"dsType"`
					Fields []struct {
						Func string `json:"func"`
						Name string `json:"name"`
					} `json:"fields"`
					GroupBy []struct {
						Params []string `json:"params"`
						Type   string   `json:"type"`
					} `json:"groupBy"`
					GroupByTags  []interface{} `json:"groupByTags"`
					Interval     string        `json:"interval"`
					Measurement  string        `json:"measurement"`
					OrderByTime  string        `json:"orderByTime"`
					Policy       string        `json:"policy"`
					Query        string        `json:"query"`
					RefID        string        `json:"refId"`
					ResultFormat string        `json:"resultFormat"`
					Select       [][]struct {
						Params []string `json:"params"`
						Type   string   `json:"type"`
					} `json:"select"`
					Tags []struct {
						Key      string `json:"key"`
						Operator string `json:"operator"`
						Value    string `json:"value"`
					} `json:"tags"`
				} `json:"model"`
				Params []string `json:"params"`
			} `json:"query"`
			Reducer struct {
				Params []interface{} `json:"params"`
				Type   string        `json:"type"`
			} `json:"reducer"`
			Type string `json:"type"`
		} `json:"conditions"`
		ExecutionErrorState string `json:"executionErrorState"`
		For                 string `json:"for"`
		Frequency           string `json:"frequency"`
		Handler             int    `json:"handler"`
		Name                string `json:"name"`
		NoDataState         string `json:"noDataState"`
		Notifications       []struct {
			UID string `json:"uid"`
		} `json:"notifications"`
	} `json:"Settings"`
}

type Annotations []Annotation

type Annotation struct {
	ID           int           `json:"id"`
	AlertID      int           `json:"alertId"`
	AlertName    string        `json:"alertName"`
	DashboardID  int           `json:"dashboardId"`
	DashboardUID string        `json:"dashboardUID"`
	PanelID      int           `json:"panelId"`
	UserID       int           `json:"userId"`
	NewState     string        `json:"newState"`
	PrevState    string        `json:"prevState"`
	Created      int64         `json:"created"`
	Updated      int64         `json:"updated"`
	Time         int64         `json:"time"`
	TimeEnd      int64         `json:"timeEnd"`
	Text         string        `json:"text"`
	Tags         []interface{} `json:"tags"`
	Login        string        `json:"login"`
	Email        string        `json:"email"`
	AvatarURL    string        `json:"avatarUrl"`
	Data         struct {
		EData EvalData
	} `json:"data"`
}

type EvalData struct {
	EvalMatches []struct {
		Metric string      `json:"metric"`
		Tags   interface{} `json:"tags"`
		Value  float64     `json:"value"`
	} `json:"evalMatches"`
}

// NewGrafanaClient создает новый экземпляр клиента
func NewGrafanaClient(baseURL string, auth string, start, end time.Time, logs *logger.LogWriter, debug bool) *GrafanaClient {
	return &GrafanaClient{
		baseURL:    baseURL,
		auth:       auth,
		client:     &http.Client{Timeout: 30 * time.Second},
		logs:       logs,
		debug:      debug,
		start:      start,
		end:        end,
		Timeperiod: "&from=" + fmt.Sprintf("%d", start.Unix()) + "000&to=" + fmt.Sprintf("%d", end.Unix()) + "000",
	}
}

func (p *GrafanaClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *GrafanaClient) GetImage(path, Name string) (string, error) {

	resp, err := http.NewRequest("GET", p.baseURL+p.Timeperiod, nil)
	if err != nil {
		return "", err
	}
	resp.Header.Add("Authorization", p.auth)
	resp.Header.Add("Content-Type", "image/jpeg")
	rsp, err := p.client.Do(resp)

	if err != nil {
		return "", err
	}

	contentype := rsp.Header["Content-Type"][0]

	//  проверяем получение картинки, статус 200
	if rsp.StatusCode == http.StatusOK {
		p.logs.ProcessDebug("Request image success " + Name)

		var n io.Reader
		// io.Copy(ioutil.Discard, rsp.Body)
		nn, err := ioutil.ReadAll(rsp.Body)
		n = bytes.NewReader(nn)

		if err != nil {
			return "", err
		}

		// open a file for writing
		file, err := os.Create(path + Name + ".png")
		if err != nil {
			return "", err
		}

		//  Use io.Copy to just dump the response body to the file. This supports huge files
		_, err = io.Copy(file, n)
		if err != nil {
			return "", err
		}

		defer rsp.Body.Close()
		defer file.Close()
	} else {
		return contentype, fmt.Errorf("grafana API returned status %d", rsp.StatusCode)
	}

	return contentype, nil
}

func (p *GrafanaClient) GetAlertStatus(alertID int) (bool, error) {

	tmp, err := p.GetAlert(alertID)

	if err != nil {
		return false, err
	}

	if tmp.State == "alerting" {
		return true, nil
	} else {
		return false, nil
	}
}

func (p *GrafanaClient) GetAlert(alertID int) (Alert, error) {
	parsedURL, err := url.Parse(p.baseURL)
	if err != nil {
		return Alert{}, fmt.Errorf("GetAlerts error parsing %w", err)
	}

	url := fmt.Sprintf("%s://%s/api/alerts/%d", parsedURL.Scheme, parsedURL.Host, alertID)

	resp, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Alert{}, fmt.Errorf("GetAlerts error request %w", err)
	}
	resp.Header.Add("Authorization", p.auth)
	resp.Header.Set("Content-Type", "application/json")
	rsp, err := p.client.Do(resp)

	if err != nil {
		return Alert{}, fmt.Errorf("GetAlerts error %w", err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rsp.Body)
		return Alert{}, fmt.Errorf("API error: %s, body: %s", rsp.Status, string(body))
	}

	var tmp Alert
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&tmp)
	if err != nil {
		p.logs.ProcessDebug(rsp.Body)
		return Alert{}, err
	}

	return tmp, nil
}

func (p *GrafanaClient) GetHistAlerts(alertID int) (bool, string, error) {
	if p.baseURL == "" {
		return false, "", fmt.Errorf("grafana: base URL is empty")
	}

	u, err := url.Parse(p.baseURL)
	if err != nil {
		return false, "", fmt.Errorf("grafana: failed to parse base URL: %w", err)
	}

	gurl := u.Scheme + "://" + u.Host + "/api/annotations?type=alert" + p.Timeperiod

	tmpalert, err := p.GetAlert(alertID)
	if err != nil {
		p.logs.ProcessError(err)
	}

	if tmpalert.DashboardID != 0 {
		gurl += "&dashboardId=" + fmt.Sprint(tmpalert.DashboardID)
	}
	if tmpalert.PanelID != 0 {
		gurl += "&panelId=" + fmt.Sprint(tmpalert.PanelID)
	}

	req, err := http.NewRequest("GET", gurl, nil)
	if err != nil {
		return false, "", fmt.Errorf("GetHistAlerts error request %w", err)
	}
	req.Header.Add("Authorization", p.auth)
	req.Header.Set("Content-Type", "application/json")
	resp, err := p.client.Do(req)

	if err != nil {
		return false, "", fmt.Errorf("GetHistAlerts error %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, "", fmt.Errorf("API error: %s, body: %s", resp.Status, string(body))
	}

	var tmp Annotations
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&tmp)
	if err != nil {
		p.logs.ProcessDebug(resp.Body)
		return false, "", err
	}

	var alerting bool
	var txt string
	for _, i := range tmp {
		if i.AlertID == alertID && i.NewState == "alerting" {
			alerting = true
			txt += fmt.Sprintf("Date the alert was issued %s \n", convertUnixMillisToTime(i.Created).Format("15:04:05"))
		}
	}

	return alerting, txt, nil
}

// Вспомогательная функция для получения первого непустого значения из списка ключей
func getFirstNonEmpty(values url.Values, keys ...string) string {
	for _, key := range keys {
		if value := values.Get(key); value != "" {
			return value
		}
	}
	return ""
}

func convertUnixMillisToTime(millis int64) time.Time {
	// Unix time в миллисекундах -> время
	return time.Unix(millis/1000, (millis%1000)*int64(time.Millisecond))
}
