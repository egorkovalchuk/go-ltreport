package allure

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/google/uuid"
)

type Allure struct {
	on      bool
	logFunc *logger.LogWriter
}

type AllureResult struct {
	UUID        string             `json:"uuid"`
	Name        string             `json:"name"`
	FullName    string             `json:"fullName"`
	HistoryID   string             `json:"historyId"`
	Status      string             `json:"status"`
	Stage       string             `json:"stage"`
	Start       int64              `json:"start"`
	Stop        int64              `json:"stop"`
	Steps       []AllureStep       `json:"steps"`
	Labels      []AllureLabel      `json:"labels"`
	Links       []AllureLink       `json:"links"`
	Attachments []AllureAttachment `json:"attachments"`
	Parameters  []AllureParameter  `json:"parameters"`
	Description string             `json:"description,omitempty"`
}

type AllureStep struct {
	Name        string             `json:"name"`
	Status      string             `json:"status"`
	Start       int64              `json:"start"`
	Stop        int64              `json:"stop"`
	Steps       []AllureStep       `json:"steps"`
	Attachments []AllureAttachment `json:"attachments"`
	Parameters  []AllureParameter  `json:"parameters"`
}

// AllureParameter - структура параметра
type AllureParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AllureAttachment struct {
	Name    string `json:"name"`
	Source  string `json:"source"`            // UUID названия файла
	Type    string `json:"type"`              // MIME type
	Content string `json:"content,omitempty"` // base64 encoded content
}

type AllureLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AllureLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

func NewAllure(on bool, logFunc *logger.LogWriter) *Allure {
	return &Allure{
		on:      on,
		logFunc: logFunc,
	}
}

func (a *Allure) CreateAllureReport(name, fullname, historyID, status, stage string, start, stop int64, labels []AllureLabel, parameters []AllureParameter, links []AllureLink) {
	if !a.on {
		return
	}
	testUUID := uuid.New().String()
	result := AllureResult{
		UUID:      testUUID,
		Name:      name,
		FullName:  fullname,
		HistoryID: historyID,
		Status:    status,
		Stage:     stage,
		Start:     start,
		Stop:      stop,
		Links:     links,
		Steps: []AllureStep{
			{
				Name:   "Checked parameter",
				Status: "passed",
				Start:  start,
				Stop:   stop,
			},
		},
		Labels:     labels,
		Parameters: parameters,
	}

	// Сохраняем в файл
	a.saveAllureResult(result)
}

func (a *Allure) createAttachment(name, mimeType string, content []byte) AllureAttachment {
	attachmentUUID := uuid.New().String()
	return AllureAttachment{
		Name:    name,
		Source:  attachmentUUID + "-attachment", // Allure ожидает такой формат
		Type:    mimeType,
		Content: base64.StdEncoding.EncodeToString(content), // base64 кодирование
	}
}

func (a *Allure) saveAllureResult(result AllureResult) {

	file, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		a.logFunc.ProcessError("Error create allure json " + err.Error())
	}
	a.createOutputDir()
	err = os.WriteFile("allure-results/"+result.UUID+"-result.json", file, 0644)
	if err != nil {
		a.logFunc.ProcessError("Error create allure json " + err.Error())
	}
}

func (a *Allure) createOutputDir() {
	isExists, err := a.exists("allure-results")
	if err != nil {
		a.logFunc.ProcessError("Error create allure json " + err.Error())
	}

	if !isExists {
		_ = os.MkdirAll("allure-results", os.ModePerm)
	}

}

func (a *Allure) exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
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
		a.logFunc.ProcessError("Error create allure json " + err.Error())
	}
	a.createOutputDir()
	// Сохраняем файл с правильным именем (как в Source)
	err = os.WriteFile("allure-results/"+attachment.Source, content, 0644)
	if err != nil {
		a.logFunc.ProcessError("Error create allure json " + err.Error())
	}

}
