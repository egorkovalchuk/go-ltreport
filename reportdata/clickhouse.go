package reportdata

import (
	"encoding/json"
	"log"
	"net/http"
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

func (p *ClickHouseJson) JsonClickHouseParse(resp *http.Response, name string) error {

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&p)

	if err != nil {
		log.Println(err)
		return err
	}

	var tmp []MetaStruct

	for _, i := range p.Meta {
		for _, j := range p.Data {
			i.Len = max(len(j[i.Name].(string)), i.Len)
		}
		i.Len = max(len(i.Name), i.Len)
		tmp = append(tmp, MetaStruct{Name: i.Name, Type: i.Type, Len: i.Len})
	}

	p.Meta = tmp
	p.ColunmLen = len(p.Meta)
	p.Name = name

	return nil

}

func (p *ClickHouseJson) DecodeResult() {

}
