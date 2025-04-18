package reportdata

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// GrafanaClient представляет клиент для работы с Grafana
type GrafanaClient struct {
	baseURL string
	auth    string
	client  *http.Client
	logFunc func(string, interface{})
	debug   bool
}

// NewGrafanaClient создает новый экземпляр клиента
func NewGrafanaClient(baseURL string, auth string, logFunc func(string, interface{}), debug bool) *GrafanaClient {
	return &GrafanaClient{
		baseURL: baseURL,
		auth:    auth,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (p *GrafanaClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *GrafanaClient) ProcessDebug(t interface{}) {
	if p.debug {
		p.logFunc("DEBUG", t)
	}
}

func (p *GrafanaClient) GetImage(Name string) (string, error) {

	resp, err := http.NewRequest("GET", p.baseURL, nil)
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
		p.logFunc("INFO", "Request image success")

		var n io.Reader
		// io.Copy(ioutil.Discard, rsp.Body)
		nn, err := ioutil.ReadAll(rsp.Body)
		n = bytes.NewReader(nn)

		if err != nil {
			return "", err
		}

		// open a file for writing
		file, err := os.Create(Name + ".png")
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

	}

	return contentype, nil
}
