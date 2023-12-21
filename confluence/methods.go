package confluence

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// режим Дебага
var DebugFlag = false
var LogFlag = false

//Запись в лог при включенном дебаге
func processdebug(logtext interface{}) {
	if DebugFlag {
		if LogFlag {
			log.Println(logtext)
		} else {
			fmt.Println(logtext)
		}
	}
}

//Вывод в лог
func processlog(logtext interface{}) {
	if LogFlag {
		log.Println(logtext)
	} else {
		fmt.Println(logtext)
	}
}

// Установить переменную дебага
func SetDebug(state bool) {
	DebugFlag = state
}

// Слямзино и переделано
// Copy Virtomize/confluence-go-api
// Конструктор
func NewAPI(location string, username string, password string, token string, proxyurl string) (*API, error) {
	if len(location) == 0 {
		return nil, errors.New("url empty")
	}

	u, err := url.ParseRequestURI(location)

	if err != nil {
		return nil, err
	}

	a := new(API)
	a.Url = u
	a.password = password
	a.username = username
	a.token = token

	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		IdleConnTimeout:       30 * time.Second}

	if proxyurl != "" {
		proxyUrl, _ := url.Parse(proxyurl)
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	a.client = &http.Client{Transport: tr}

	return a, nil
}

func (confl *API) GetContent(id string, param ContentQuery) (*ConflType, error) {
	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/" + id)

	if err != nil {
		return nil, errors.New("Url generation error")
	}

	log.Println("Load " + ep.String())
	ep.RawQuery = addContentQueryParams(param).Encode()

	req, err := http.NewRequest("GET", ep.String(), nil)
	if err != nil {
		return nil, errors.New("Error creating request to " + ep.String())
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := confl.Request(req)
	if err != nil {
		return nil, err
	}

	JsonCont, err := confl.GetJson(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonCont, nil

	//return res, nil

}

func (confl *API) GetContentChildPage(id string, param ContentQuery) (*ConflTypeA, error) {
	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/" + id + "/child/page")

	if err != nil {
		return nil, errors.New("Url generation error")
	}

	log.Println("Load " + ep.String())
	ep.RawQuery = addContentQueryParams(param).Encode()

	req, err := http.NewRequest("GET", ep.String(), nil)
	if err != nil {
		return nil, errors.New("Error creating request to " + ep.String())
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := confl.Request(req)
	if err != nil {
		return nil, err
	}

	JsonCont, err := confl.GetJsonA(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonCont, nil

}

func (confl *API) GetJson(content []byte) (*ConflType, error) {

	var JsonContent ConflType
	err := json.Unmarshal(content, &JsonContent)
	if err != nil {
		return nil, err
	}
	return &JsonContent, nil

}

//Создание новой страницы
func (confl *API) CreateContent(data *ConflCreateType) (*ConflType, error) {

	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/")

	if err != nil {
		return nil, errors.New("Url generation error")
	}

	log.Println("Load " + ep.String())

	var body io.Reader
	if data != nil {
		js, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = strings.NewReader(string(js))
	}

	req, err := http.NewRequest("POST", ep.String(), body)
	if err != nil {
		return nil, errors.New("Error creating request to " + ep.String())
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := confl.Request(req)
	if err != nil {
		return nil, err
	}

	JsonConC, err := confl.GetJson(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonConC, nil

	//return res, nil

}

func (confl *API) UploadAttachment(id string, attachmentName string, attachment io.Reader) (*ConflTypeA, error) {

	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/" + id + "/child/attachment")
	if err != nil {
		return nil, err
	}

	res, err := confl.SendContentAttachmentRequest(ep, attachmentName, attachment, map[string]string{})

	if err != nil {
		return nil, err
	}

	JsonCon, err := confl.GetJsonA(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonCon, nil

}

func (confl *API) UpdateAttachment(id string, attachmentName string, attachid string, attachment io.Reader) (*ConflTypeA, error) {

	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/" + id + "/child/attachment/" + attachid + "/data")
	if err != nil {
		return nil, err
	}

	res, err := confl.SendContentAttachmentRequest(ep, attachmentName, attachment, map[string]string{})
	if err != nil {
		return nil, err
	}

	JsonCon, err := confl.GetJsonA(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonCon, nil

}

func (confl *API) GetAttachments(id string) (*ConflTypeA, error) {
	ep, err := url.ParseRequestURI(confl.Url.String() + "/rest/api/content/" + id + "/child/attachment")

	if err != nil {
		return nil, errors.New("Url generation error")
	}

	log.Println("Load " + ep.String())

	req, err := http.NewRequest("GET", ep.String(), nil)
	if err != nil {
		return nil, errors.New("Error creating request to " + ep.String())
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := confl.Request(req)
	if err != nil {
		return nil, err
	}

	JsonCon, err := confl.GetJsonA(res)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return JsonCon, nil

}

func (confl *API) GetJsonA(content []byte) (*ConflTypeA, error) {

	var JsonContent ConflTypeA
	err := json.Unmarshal(content, &JsonContent)
	if err != nil {
		return nil, err
	}
	return &JsonContent, nil

}
