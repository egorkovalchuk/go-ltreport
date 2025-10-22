package allure

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (ar *AllureResult) FilishedFailed() {
	ar.Status = "failed"
	ar.Stage = "finished"
}

func (ar *AllureResult) FilishedPassed() {
	ar.Status = "passed"
	ar.Stage = "finished"
}

func (ar *AllureResult) FilishedBroken() {
	ar.Status = "broken"
	ar.Stage = "finished"
}

func (ar *AllureResult) ArrayToLabelRoot(txt, label string) {
	headers := strings.Split(txt, ";")
	for _, i := range headers {
		before, after, found := strings.Cut(i, ":")
		if found {
			ar.Labels = append(ar.Labels, AllureLabel{Name: before, Value: after})
		} else {
			ar.Labels = append(ar.Labels, AllureLabel{Name: label, Value: i})
		}
	}
}

func (ar *AllureResult) GetValueLabel(name string) string {
	for _, i := range ar.Labels {
		if i.Name == name {
			return i.Value
		}
	}
	return "Unknown"
}

func (ar *AllureResult) AddStep(name, status string, start, end time.Time) {
	ar.Steps = append(ar.Steps, AllureStep{Name: name, Status: status, Start: start.UnixMilli(), Stop: end.UnixMilli()})
}

func (ar *AllureResult) AddLabel(name, value string) {
	ar.Labels = append(ar.Labels, AllureLabel{Name: name, Value: value})
}

func (ar *AllureResult) AddStepWithParam(name, status string, start, end time.Time, p []AllureParameter) {
	ar.Steps = append(ar.Steps, AllureStep{Name: name, Status: status, Start: start.UnixMilli(), Stop: end.UnixMilli(), Parameters: p})
}

func (ar *AllureResult) AddStepChild(name, status string, start, end time.Time, p []AllureParameter, s []AllureStep, a []AllureAttachment, ars *AllureStep) {
	ars.Steps = append(ars.Steps, AllureStep{Name: name, Status: status, Start: start.UnixMilli(), Stop: end.UnixMilli(), Parameters: p, Steps: s, Attachments: a})
}

func (ar *AllureResult) AddLink(name, url, typ string) {
	ar.Links = append(ar.Links, AllureLink{Name: name, URL: url, Type: typ})
}

func (ar *AllureResult) AddAttach(name string, tp AttachmentType, filename string, content []byte) error {
	attachmentUUID := uuid.New().String()
	if len(content) > 0 {
		ar.Attachments = append(ar.Attachments, AllureAttachment{Name: name, Source: attachmentUUID + "-attachment", Type: string(tp), Content: base64.StdEncoding.EncodeToString(content)})
	}
	if filename != "" {
		// Проверяем существование файла
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filename)
		}

		// Читаем файл
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		// Определяем MIME type по расширению
		attachmentType := getAttachmentType(filename)

		// Получаем имя файла для отображения в отчете
		fileName := filepath.Base(filename)

		ar.Attachments = append(ar.Attachments, AllureAttachment{Name: fileName, Source: attachmentUUID + "-attachment", Type: string(attachmentType), Content: base64.StdEncoding.EncodeToString(content)})
	}
	return nil
}

func (ar *AllureResult) ArrayToParamRoot(txt string) {
	headers := strings.Split(txt, ";")
	for _, i := range headers {
		before, after, found := strings.Cut(i, ":")
		if found {
			ar.Parameters = append(ar.Parameters, AllureParameter{Name: before, Value: after})
		}
	}
}

func (ar *AllureResult) ArrayToParam(txt string) []AllureParameter {
	headers := strings.Split(txt, ";")
	var tmp []AllureParameter
	for _, i := range headers {
		before, after, found := strings.Cut(i, ":")
		if found {
			tmp = append(tmp, AllureParameter{Name: before, Value: after})
		}
	}
	return tmp
}
