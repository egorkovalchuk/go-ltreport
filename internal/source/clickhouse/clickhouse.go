package clickhouse

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"
)

// NewCHClient создает новый экземпляр клиента
func NewCHClient(baseURL string, user string, pass string, start, end time.Time, logs *logger.LogWriter, debug bool) *CHClient {
	return &CHClient{
		baseURL:    baseURL,
		user:       user,
		pass:       pass,
		client:     &http.Client{Timeout: 120 * time.Second},
		logs:       logs,
		debug:      debug,
		start:      start,
		end:        end,
		Timeperiod: " timestamp>=toDateTime('" + start.UTC().Format("2006-01-02 15:04:05") + "') and timestamp <=toDateTime('" + end.UTC().Format("2006-01-02 15:04:05") + "') ",
	}
}

func (p *CHClient) GetSql(DBname string, sql string, name string) (ClickHouseJson, error) {

	resp, err := http.NewRequest("GET", p.baseURL, nil)
	if err != nil {
		return ClickHouseJson{}, fmt.Errorf("GetSql request failed: %w", err)
	}
	p.logs.ProcessDebug(strings.Replace(sql, "{timestamp}", p.Timeperiod, 1) + " FORMAT JSONStrings")

	resp.SetBasicAuth(p.user, p.pass)
	resp.Header.Add("Content-Type", "application/json")
	resp.Header.Add("X-ClickHouse-Progress", "1")
	resp.Header.Add("X-ClickHouse-Database", DBname)
	resp.Header.Add("User-Agent", "go-LT-Report")
	resp.Body = io.NopCloser(strings.NewReader(strings.Replace(sql, "{timestamp}", p.Timeperiod, 1) + " FORMAT JSONStrings"))

	rsp, err := p.client.Do(resp)
	if err != nil {
		p.logs.ProcessError(fmt.Errorf("GetSql request failed: %w", err))
		return ClickHouseJson{}, fmt.Errorf("GetSql request failed: %w", err)
	}

	if rsp.StatusCode == http.StatusOK {
		p.logs.ProcessInfo("Query ClickHouse succes ")
	} else {
		p.logs.ProcessError(fmt.Errorf("ClickHouse API returned status %d", rsp.StatusCode))
		p.logs.ProcessDebug(resp)
		return ClickHouseJson{}, fmt.Errorf("ClickHouse API returned status %d", rsp.StatusCode)
	}

	var clkhouse ClickHouseJson
	err = clkhouse.JsonClickHouseParse(rsp, name)
	if err != nil {
		p.logs.ProcessError(fmt.Errorf("GetSql JsonClickHouseParse failed: %w", err))
		return ClickHouseJson{}, fmt.Errorf("GetSql JsonClickHouseParse failed: %w", err)
	}

	p.client.CloseIdleConnections()

	return clkhouse, nil
}

func (p *CHClient) Close() {
	p.client.CloseIdleConnections()
}

func (p *ClickHouseJson) JsonClickHouseParse(resp *http.Response, name string) error {

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err := decoder.Decode(&p)

	if err != nil {
		return fmt.Errorf("CH: %w", err)
	}

	var tmp []MetaStruct

	for _, i := range p.Meta {
		for _, j := range p.Data {
			i.Len = reportdata.MaxInt(len(j[i.Name].(string)), i.Len)
		}
		i.Len = reportdata.MaxInt(len(i.Name), i.Len)
		tmp = append(tmp, MetaStruct{Name: i.Name, Type: i.Type, Len: i.Len})
	}

	p.Meta = tmp
	p.ColunmLen = len(p.Meta)
	p.Name = name

	return nil

}

func (p *ClickHouseJson) DecodeResult() {

}

// Вычислем максимальные длины строки
// Пересчет
func (p *ClickHouseJson) Lens() map[string]int {
	tmplen := make(map[string]int, len(p.Meta))

	for _, m := range p.Meta {
		tmplen[m.Name] = len(m.Name)
	}

	for _, m := range p.Data {
		for key, value := range tmplen {

			tmplen[key] = reportdata.MaxInt(len(m[key].(string)), value)

			var strValue string
			switch v := m[key].(type) {
			case string:
				strValue = v
			case fmt.Stringer:
				strValue = v.String()
			case int, float64, bool:
				strValue = fmt.Sprintf("%v", v)
			default:
				continue // пропускаем неподдерживаемые типы
			}

			currentLength := len(strValue)
			if currentMax, exists := tmplen[key]; !exists || currentLength > currentMax {
				tmplen[key] = currentLength
			}
		}
	}
	return tmplen
}

func (p *ClickHouseJson) RoundToPrecision(precision int) {
	for i, m := range p.Data {
		for key, value := range m {
			if num, ok := reportdata.ConvIntefaceFloat64(value); ok {
				p.Data[i][key] = fmt.Sprint(reportdata.RoundToPrecision(num, 4))
			}
		}
	}
}
