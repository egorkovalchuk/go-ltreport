package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

// NewGrafanaClient создает новый экземпляр клиента
func NewGrafanaClient(baseURL string, auth string, start, end time.Time, logs *logger.LogWriter, debug bool) *GrafanaClient {
	tmp := &GrafanaClient{
		baseURL:    baseURL,
		auth:       auth,
		client:     &http.Client{Timeout: 30 * time.Second},
		logs:       logs,
		debug:      debug,
		start:      start,
		end:        end,
		Timeperiod: "&from=" + fmt.Sprintf("%d", start.Unix()) + "000&to=" + fmt.Sprintf("%d", end.Unix()) + "000",
	}

	tmp.version = tmp.GetVersionGrafana()
	return tmp
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
		// io.Copy(io.Discard, rsp.Body)
		nn, err := io.ReadAll(rsp.Body)
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

func (p *GrafanaClient) GetAlert(alertID any) (requestAnno, error) {
	parsedURL, err := url.Parse(p.baseURL)
	if err != nil {
		return requestAnno{}, fmt.Errorf("GetAlerts error parsing %w", err)
	}

	url := ""
	alertIDtmp, err := alertToString(alertID)
	if err != nil {
		return requestAnno{}, fmt.Errorf("GetAlerts error type: %w", err)
	}

	if p.version < 11 {
		url = fmt.Sprintf("%s://%s/api/alerts/%s", parsedURL.Scheme, parsedURL.Host, alertIDtmp)
	} else {
		url = fmt.Sprintf("%s://%s/api/v1/provisioning/alert-rules/%s", parsedURL.Scheme, parsedURL.Host, alertIDtmp)
	}

	resp, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return requestAnno{}, fmt.Errorf("GetAlerts error request %w", err)
	}
	resp.Header.Add("Authorization", p.auth)
	resp.Header.Set("Content-Type", "application/json")
	rsp, err := p.client.Do(resp)

	if err != nil {
		return requestAnno{}, fmt.Errorf("GetAlerts error %w", err)
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rsp.Body)
		return requestAnno{}, fmt.Errorf("API error: %s, body: %s", rsp.Status, string(body))
	}

	tmprequestAnno := requestAnno{}
	if p.version < 11 {
		var tmp AlertV10
		decoder := json.NewDecoder(rsp.Body)
		err = decoder.Decode(&tmp)
		if err != nil {
			p.logs.ProcessDebug(rsp.Body)
			return requestAnno{}, err
		}
		tmprequestAnno.DashID = fmt.Sprint(tmp.DashboardID)
		tmprequestAnno.PanelID = fmt.Sprint(tmp.PanelID)
	} else {
		var tmp AlertV11
		decoder := json.NewDecoder(rsp.Body)
		err = decoder.Decode(&tmp)
		if err != nil {
			p.logs.ProcessDebug(rsp.Body)
			return requestAnno{}, err
		}
		tmprequestAnno.DashUID = fmt.Sprint(tmp.Annotations.DashboardUid)
		tmprequestAnno.PanelID = fmt.Sprint(tmp.Annotations.PanelID)
	}

	return tmprequestAnno, nil
}

func (p *GrafanaClient) GetHistAlerts(alertID any) (bool, string, error) {
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

	if tmpalert.DashID != "" {
		gurl += "&dashboardId=" + fmt.Sprint(tmpalert.DashID)
	}
	if tmpalert.DashUID != "" {
		gurl += "&dashboardUid=" + fmt.Sprint(tmpalert.DashID)
	}
	if tmpalert.PanelID != "" {
		gurl += "&panelId=" + fmt.Sprint(tmpalert.PanelID)
	}

	p.logs.ProcessDebug("Alerts url: " + gurl)

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
		if strings.ToLower(i.NewState) == "alerting" {
			alerting = true
			txt += fmt.Sprintf("Date the alert was issued %s \n", convertUnixMillisToTime(i.Created).Format("15:04:05"))
		}
	}

	return alerting, txt, nil
}

func (p *GrafanaClient) GetVersionGrafana() (version int) {
	version = 0

	parsedURL, err := url.Parse(p.baseURL)
	if err != nil {
		return version
	}

	url := fmt.Sprintf("%s://%s/api/health", parsedURL.Scheme, parsedURL.Host)

	resp, err := http.NewRequest("GET", url, nil)
	if err != nil {
		p.logs.ProcessInflux(fmt.Errorf("GetVersionGrafana error request %w", err))
		return 0
	}
	resp.Header.Add("Authorization", p.auth)
	resp.Header.Set("Content-Type", "application/json")
	rsp, err := p.client.Do(resp)

	if err != nil {
		p.logs.ProcessInflux(fmt.Errorf("GetVersionGrafana error request %w", err))
		return 0
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rsp.Body)
		p.logs.ProcessInflux(fmt.Errorf("GetVersionGrafana API error: %s, body: %s", rsp.Status, string(body)))
		return 0
	}

	var tmp health
	decoder := json.NewDecoder(rsp.Body)
	err = decoder.Decode(&tmp)
	if err != nil {
		p.logs.ProcessDebug(rsp.Body)
		return 0
	}
	p.logs.ProcessDebug("Grafana version: " + tmp.Version)
	return getMajorVersion(tmp.Version)
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

func getMajorVersion(version string) int {
	parts := strings.Split(version, ".")
	if len(parts) == 0 {
		return 0
	}
	tmp, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	return tmp
}

func alertToString(AlertId any) (string, error) {
	switch {
	case AlertId == nil:
		return "", fmt.Errorf("AlertID is missing")
	default:
		// Дополнительная проверка типа
		switch v := AlertId.(type) {
		case int:
			return strconv.Itoa(v), nil
		case string:
			return v, nil
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case int64:
			return strconv.FormatInt(v, 10), nil
		default:
			return "", fmt.Errorf("Unknown type AlertID(%T)", AlertId)
		}
	}
}
