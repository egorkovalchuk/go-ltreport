package reportdata

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Result of Query.
type Result interface {
	//	DecodeResult(r *Reader, version int, b Block) error
}

type MetaStruct struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Len  int
}
type ClickHouseJson struct {
	ColunmLen int
	Name      string
	Meta      []MetaStruct `json:"meta"`
	Data      []map[string]interface {
	} `json:"data"`
	Rows       int `json:"rows"`
	Statistics struct {
		Elapsed    float32 `json:"elapsed"`
		Rows_read  float32 `json:"rows_read"`
		Bytes_read float32 `json:"bytes_read"`
	} `json:"statistics"`
}

// CHClient представляет клиент для работы с Grafana
type CHClient struct {
	baseURL string
	user    string
	pass    string
	client  *http.Client
	logFunc func(string, interface{})
	debug   bool
}

// NewCHClient создает новый экземпляр клиента
func NewCHClient(baseURL string, user string, pass string, logFunc func(string, interface{}), debug bool) *CHClient {
	return &CHClient{
		baseURL: baseURL,
		user:    user,
		pass:    pass,
		client:  &http.Client{Timeout: 30 * time.Second},
		logFunc: logFunc,
		debug:   debug,
	}
}

func (p *CHClient) GetSql(DBname string, sql string, name string, timeperiod string) (ClickHouseJson, error) {

	resp, err := http.NewRequest("GET", p.baseURL, nil)
	if err != nil {
		return ClickHouseJson{}, fmt.Errorf("GetSql request failed: %v", err)
	}
	p.ProcessDebug(strings.Replace(sql, "{timestamp}", timeperiod, 1) + " FORMAT JSONStrings")

	resp.SetBasicAuth(p.user, p.pass)
	resp.Header.Add("Content-Type", "application/json")
	resp.Header.Add("X-ClickHouse-Progress", "1")
	resp.Header.Add("X-ClickHouse-Database", DBname)
	resp.Header.Add("User-Agent", "go-LT-Report")
	resp.Body = ioutil.NopCloser(strings.NewReader(strings.Replace(sql, "{timestamp}", timeperiod, 1) + " FORMAT JSONStrings"))

	rsp, err := p.client.Do(resp)
	if err != nil {
		return ClickHouseJson{}, fmt.Errorf("GetSql request failed: %v", err)
	}

	if rsp.StatusCode == http.StatusOK {
		p.logFunc("INFO", "Query ClickHouse succes ")
	} else {
		return ClickHouseJson{}, err
	}

	var clkhouse ClickHouseJson
	err = clkhouse.JsonClickHouseParse(rsp, name)
	if err != nil {
		return ClickHouseJson{}, err
	}

	p.client.CloseIdleConnections()

	return clkhouse, nil
}

func (p *CHClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *CHClient) ProcessDebug(t interface{}) {
	if p.debug {
		p.logFunc("DEBUG", t)
	}
}

func (p *ClickHouseJson) JsonClickHouseParse(resp *http.Response, name string) error {

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&p)

	if err != nil {
		return fmt.Errorf("CH: %v", err)
	}

	var tmp []MetaStruct

	for _, i := range p.Meta {
		for _, j := range p.Data {
			i.Len = MaxInt(len(j[i.Name].(string)), i.Len)
		}
		i.Len = MaxInt(len(i.Name), i.Len)
		tmp = append(tmp, MetaStruct{Name: i.Name, Type: i.Type, Len: i.Len})
	}

	p.Meta = tmp
	p.ColunmLen = len(p.Meta)
	p.Name = name

	return nil

}

func (p *ClickHouseJson) DecodeResult() {

}
