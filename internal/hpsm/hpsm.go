package hpsm

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
)

type FSMClient struct {
	baseURL  string
	token    string
	tu       string
	user     string
	password string
	period   string
	client   *http.Client
	logFunc  *logger.LogWriter
	debug    bool
}

type Content struct {
	Content []struct {
		AvrName          string `json:"avrName"`
		BriefDescription string `json:"briefDescription"`
		Description      string `json:"description"`
		Number           string `json:"number"`
	} `json:"content"`
}

func NewFSM(baseURL, tu, token, user, password, period string, logFunc *logger.LogWriter) *FSMClient {
	return &FSMClient{
		baseURL:  baseURL,
		tu:       tu,
		token:    token,
		user:     user,
		password: password,
		logFunc:  logFunc,
		period:   period,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *FSMClient) GetIM() Content {

	resp, err := http.NewRequest("GET", f.baseURL+"?LogicalName="+f.tu+"&"+f.period, nil)

	if f.user != "" && f.password != "" {
		resp.SetBasicAuth(f.user, f.password)
	} else if f.token != "" {
		resp.Header.Add("Token", f.token)
	} else {
		f.logFunc.ProcessInfo("FSM authorization parameters are not set")
		return Content{}
	}

	resp.Header.Add("Connection", "close")
	resp.Header.Add("User-Agent", "go-LT-Report")

	f.logFunc.ProcessInfo(resp)

	rsp, err := f.client.Do(resp)
	if err != nil {
		f.logFunc.ProcessError(err)
		return Content{}
	}

	var infjson Content

	if rsp.StatusCode == http.StatusOK {
		decoder := json.NewDecoder(rsp.Body)
		decoder.UseNumber()

		err = decoder.Decode(&infjson)
		f.logFunc.ProcessInfo(infjson)
		return infjson
	} else {
		return Content{}
	}

}
