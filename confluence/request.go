package confluence

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

//Copy Virtomize/confluence-go-api
//Выполнение самого запроса
func (confl *API) Request(req *http.Request, f func(level string, logtext interface{})) ([]byte, error) {
	req.Header.Add("Accept", "application/json, */*")

	// only auth if we can auth
	if (confl.token != "") || ((confl.username != "") && (confl.password != "")) {
		confl.Auth(req)
	}

	resp, err := confl.client.Do(req)
	if err != nil {
		f("ERROR: HTTP:", req)
		return nil, err
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		f("ERROR: HTTP:", string(res))
		return nil, err
	}

	resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		return res, nil
	case http.StatusNoContent, http.StatusResetContent:
		return nil, nil
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("authentication failed")
	case http.StatusServiceUnavailable:
		return nil, fmt.Errorf("service is not available: %s", resp.Status)
	case http.StatusInternalServerError:
		return nil, fmt.Errorf("internal server error: %s", resp.Status)
	case http.StatusConflict:
		return nil, fmt.Errorf("conflict: %s", resp.Status)
	}

	return nil, fmt.Errorf("unknown response status: %s", resp.Status)
}

// SendContentAttachmentRequest sends a multipart/form-data attachment create/update request to a content
func (confl *API) SendContentAttachmentRequest(ep *url.URL, attachmentName string, attachment io.Reader, params map[string]string, f func(level string, logtext interface{})) ([]byte, error) {
	// setup body for mulitpart file, adding minorEdit option
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("minorEdit", "true")
	part, err := writer.CreateFormFile("file", attachmentName)
	if err != nil {
		return nil, err
	}

	// add attachment to body
	_, err = io.Copy(part, attachment)
	if err != nil {
		return nil, err
	}

	// add any other params
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	//clean up multipart form writer
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", ep.String(), body) // will always be put
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Atlassian-Token", "nocheck") // required by api
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// https://developer.atlassian.com/cloud/confluence/rest/#api-api-content-id-child-attachment-put

	res, err := confl.Request(req, f)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//Аутентификация по логину паролю
//Надо добавить по токену
func (confl *API) Auth(req *http.Request) {
	//Supports unauthenticated access to confluence:
	//if username and token are not set, do not add authorization header
	if confl.username != "" && confl.password != "" {
		req.SetBasicAuth(confl.username, confl.password)
	} else if confl.token != "" {
		req.Header.Set("Authorization", "Bearer "+confl.token)
	}
}

func addContentQueryParams(query ContentQuery) *url.Values {

	data := url.Values{}
	if len(query.Expand) != 0 {
		data.Set("expand", strings.Join(query.Expand, ","))
	}
	//get specific version
	if query.Version != 0 {
		data.Set("version", strconv.Itoa(query.Version))
	}
	if query.Limit != 0 {
		data.Set("limit", strconv.Itoa(query.Limit))
	}
	if query.OrderBy != "" {
		data.Set("orderby", query.OrderBy)
	}
	if query.PostingDay != "" {
		data.Set("postingDay", query.PostingDay)
	}
	if query.SpaceKey != "" {
		data.Set("spaceKey", query.SpaceKey)
	}
	if query.Start != 0 {
		data.Set("start", strconv.Itoa(query.Start))
	}
	if query.Status != "" {
		data.Set("status", query.Status)
	}
	if query.Title != "" {
		data.Set("title", query.Title)
	}
	if query.Trigger != "" {
		data.Set("trigger", query.Trigger)
	}
	if query.Type != "" {
		data.Set("type", query.Type)
	}
	return &data
}
