package reportdata

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

type PrometheusResponse struct {
	Status    string `json:"status"`
	IsPartial bool   `json:"isPartial"`
	Data      struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Host string `json:"host"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (p *PrometheusResponse) JsonPrometheusParse(resp *http.Response) error {

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&p)

	if err != nil {
		log.Println(err)
	}

	if p.Status != "success" {
		log.Printf("Expected exactly one result in response, got %s", p.Status)
		return errors.New("Expected exactly one result in response")
	}

	if len(p.Data.Result) == 0 {
		log.Printf("Expected exactly one series in result, got %d", len(p.Data.Result))
		return errors.New("Expected exactly one series in result")
	}

	return nil

}

func (p *PrometheusResponse) JsonPrometheusFiledParse(field interface{}) SField {
	var fieldp SField

	if field != nil {
		if ff, ok := field.(string); ok {
			fieldp.ValString = ff
			if s, err := strconv.ParseFloat(ff, 64); err == nil {
				fieldp.ValFloat = s
			} else {
				fieldp.ValFloat = 0
			}
			if s, err := strconv.ParseInt(ff, 0, 64); err == nil {
				fieldp.ValInt = s
			} else {
				fieldp.ValInt = 0
			}
		} else {
			fieldp.ValFloat = 0
		}

	} else {
		fieldp.ValFloat = 0
	}

	return fieldp
}

func (p *PrometheusResponse) JsonPrometheusFiledParseFloat(field interface{}) float64 {
	fieldp := p.JsonPrometheusFiledParse(field)
	return fieldp.ValFloat
}

func (p *PrometheusResponse) JsonPrometheusFiledParseInt(field interface{}) int64 {
	fieldp := p.JsonPrometheusFiledParse(field)
	return fieldp.ValInt
}

func (p *PrometheusResponse) JsonPrometheusFiledParseString(field interface{}) string {
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValString
}
