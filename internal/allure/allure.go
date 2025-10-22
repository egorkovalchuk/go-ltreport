package allure

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Allure struct {
	on         bool
	e          chan *LogStruct
	token      string
	userToken  string
	BaseURL    string
	ProjectID  int
	HTTPClient *http.Client
	path       string
	testcount  int
}

func NewAllure(on bool, userToken, BaseURL, path string) *Allure {
	return &Allure{
		on:        on,
		userToken: userToken,
		BaseURL:   BaseURL,
		ProjectID: 367,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		path:      EnsureTrailingSeparator(path),
		e:         make(chan *LogStruct),
		testcount: 0,
	}
}

func (a *Allure) CreateAllureReport(tmp AllureResult) {
	if !a.on {
		return
	}
	result := tmp
	// Сохраняем в файл
	a.saveAllureResult(result)
	a.saveAllureAttachments(result)
}

func (a *Allure) saveAllureResult(result AllureResult) {

	file, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		a.ProcessError("Error create allure json " + err.Error())
		return
	}
	a.createOutputDir()
	err = os.WriteFile(a.path+result.UUID+"-result.json", file, 0644)
	if err != nil {
		a.ProcessError("Error create allure json " + err.Error())
		return
	}
	a.testcount++
}

func (a *Allure) saveAllureAttachments(result AllureResult) {
	// Сохраняем аттачменты из основных steps
	for _, step := range result.Steps {
		for _, attachment := range step.Attachments {
			a.saveAttachmentFile(attachment)
		}
	}

	// Сохраняем корневые аттачменты
	for _, attachment := range result.Attachments {
		a.saveAttachmentFile(attachment)
	}
}

func (a *Allure) saveAttachmentFile(attachment AllureAttachment) {
	// Декодируем base64 контент
	content, err := base64.StdEncoding.DecodeString(attachment.Content)
	if err != nil {
		a.ProcessError("Error create allure json " + err.Error())
	}
	a.createOutputDir()
	// Сохраняем файл с правильным именем (как в Source)
	err = os.WriteFile(a.path+attachment.Source, content, 0644)
	if err != nil {
		a.ProcessError("Error create allure json " + err.Error())
	}

}

func (a *Allure) Finish(zipFileName string, start, end time.Time, url string) error {
	if !a.on {
		return nil
	}
	a.createZip(zipFileName)
	a.deleteFiles()

	err := a.GetToken()
	if err != nil {
		a.ProcessError("Error get token")
		a.ProcessError(err)
		return err
	}

	id, err := a.CreateLaunch(start, end, url)
	if err != nil {
		a.ProcessError("Error create launch")
		a.ProcessError(err)
		return err
	}

	a.ProcessDebug("Launch id: " + fmt.Sprint(id))

	err = a.UploadArchive(zipFileName, id)

	return err
}

func (a *Allure) UploadArchive(zipFileName string, id int) error {

	filePath := a.path + zipFileName
	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("File not exists: %s", filePath)
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Cannot open file: %w", err)
	}
	defer file.Close()

	// Создаем multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	contentType := a.getContentType(filePath)
	filename := filepath.Base(filePath)

	// Создаем часть с явным указанием Content-Type
	part, err := writer.CreatePart(map[string][]string{
		"Content-Type":        {contentType},
		"Content-Disposition": {fmt.Sprintf(`form-data; name="archive"; filename="%s"`, filename)},
	})
	if err != nil {
		return err
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("Cannot copy error file: %w", err)
	}

	// Создаем часть для JSON данных
	jsonPart, err := writer.CreatePart(map[string][]string{
		"Content-Type":        {"application/json"},
		"Content-Disposition": {`form-data; name="info"`},
	})
	if err != nil {
		return err
	}

	_, err = jsonPart.Write([]byte("{}"))
	if err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("Error Close writer: %w", err)
	}

	// Создаем запрос
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/rs/launch/%d/upload", a.BaseURL, id), body)
	if err != nil {
		return fmt.Errorf("Error creating query: %w", err)
	}

	// Устанавливаем заголовки
	req.Header.Set("accept", "*/*")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", a.token)

	// Выполняем запрос
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error query execute: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error: %d, response: %s", resp.StatusCode, string(responseBody))
	}

	// Читаем успешный ответ
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error read response: %w", err)
	}

	var uploadResp UploadResponse
	err = json.Unmarshal(responseBody, &uploadResp)
	if err != nil {
		return err
	}
	a.ProcessDebug("Load files count: " + fmt.Sprint(uploadResp.FilesCount))

	var cnt int
	exitcount := 0
	for {
		cnt, err = a.LaunchStat(id)
		if err != nil {
			a.ProcessError(err)
			exitcount++
		} else if cnt >= a.testcount {
			err = a.CloseLaunch(id)
			if err != nil {
				a.ProcessError("Error close launch")
				a.ProcessError(err)
				return err
			}
			break
		} else if exitcount >= 100 {
			a.ProcessError("Error process files, close manualy launch")
			break
		} else {
			a.ProcessInfo(fmt.Sprintf("%d tests out of %d processed", cnt, a.testcount))
			a.ProcessInfo("Waiting process files")
			exitcount++
		}
		<-time.Tick(3 * time.Second)
	}
	a.ProcessInfo(fmt.Sprintf("%d tests out of %d processed", cnt, a.testcount))

	return nil
}

func (a *Allure) GetToken() error {
	a.ProcessDebug("Get Allure jwt token")

	// Подготавливаем данные формы
	formData := url.Values{}
	formData.Set("grant_type", "apitoken")
	formData.Set("scope", "openid")
	formData.Set("token", a.userToken)

	// Создаем запрос
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/uaa/oauth/token", a.BaseURL), bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return fmt.Errorf("Error creating query: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Expect", "")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return err
	}

	if tokenResp.TokenType == "bearer" {
		a.token = "Bearer " + tokenResp.AccessToken
		return nil
	} else {
		return nil
	}
}
