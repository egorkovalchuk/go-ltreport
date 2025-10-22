package allure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (a *Allure) CreateLaunch(start, end time.Time, url string) (int, error) {
	a.ProcessDebug("Create Launch")
	// Подготавливаем запрос на создание launch
	launchReq := &LaunchRequest{
		ProjectID: a.ProjectID,
		Name:      "Automated Report Run - " + start.Format("2006-01-02 15:04:05") + "-" + end.Format("15:04:05"),
		Autoclose: false,
		Tags:      []AllureLaunchTag{},
		Links:     []AllureLink{{Name: "Confluence", URL: url, Type: "requirement"}},
	}

	// Подготавливаем тело запроса
	jsonData, err := json.Marshal(launchReq)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/rs/launch", a.BaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", a.token)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var launchResp LaunchResponse
	err = json.Unmarshal(body, &launchResp)
	if err != nil {
		return 0, err
	}

	return launchResp.ID, nil
}

// GetLaunch получает информацию о launch
func (a *Allure) GetLaunch(launchID int) error {

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/rs/launch/%d", a.BaseURL, launchID), nil)
	if err != nil {
		return fmt.Errorf("Error creating query: %w", err)
	}
	req.Header.Set("Authorization", a.token)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var launchResp GetLaunchResponse
	err = json.Unmarshal(body, &launchResp)
	if err != nil {
		return err
	}

	return nil
}

func (a *Allure) CloseLaunch(launchID int) error {

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/rs/launch/%d/close", a.BaseURL, launchID), nil)
	if err != nil {
		return fmt.Errorf("Error creating query: %w", err)
	}
	req.Header.Set("Authorization", a.token)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (a *Allure) LaunchStat(launchID int) (int, error) {
	a.ProcessDebug("Load statistic launch")
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/rs/launch/%d/statistic", a.BaseURL, launchID), nil)
	if err != nil {
		return 0, fmt.Errorf("Error creating query: %w", err)
	}
	req.Header.Set("Authorization", a.token)

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var launchResp LaunchStatResponse
	err = json.Unmarshal(body, &launchResp)
	if err != nil {
		return 0, err
	}

	var sum int
	for _, i := range launchResp {
		sum = sum + i.Count
	}

	return sum, nil
}
