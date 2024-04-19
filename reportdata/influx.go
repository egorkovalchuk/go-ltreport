package reportdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// режим Дебага
var DebugFlag = false
var LogFlag = false

var err error

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

func JsonINfluxFiledParse(field interface{}) SField {
	var fieldp SField
	var err error

	if field != nil {
		fieldp.ValFloat, err = field.(json.Number).Float64()
		if err != nil {
			fieldp.ValFloat = 0
		}
	} else {
		fieldp.ValFloat = 0
	}

	return fieldp
}

func JsonINfluxFiledParseFloat(field interface{}) float64 {
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValFloat
}

func JsonINfluxFiledParseInt(field interface{}) int64 {
	fieldp := JsonINfluxFiledParse(field)
	return fieldp.ValInt
}

func JsonINfluxParse(resp *http.Response) (Mean, error) {

	var infjson Mean

	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()

	err = decoder.Decode(&infjson)

	if err != nil {
		log.Println(err)
		return infjson, err
	}

	if len(infjson.Results) == 0 {
		log.Printf("Expected exactly one result in response, got %d", len(infjson.Results))
		return infjson, errors.New("Expected exactly one result in response")
	}

	if len(infjson.Results[0].Series) == 0 {
		log.Printf("Expected exactly one series in result, got %d", len(infjson.Results[0].Series))
		return infjson, errors.New("Expected exactly one series in result")
	}

	return infjson, nil
}

func InfluxJmeterScenarioStatut(i [][]interface{}, statut string, cfg []JmeterQScnrFieldS) []YField {

	var LTTest_yfield []YField
	var LTTest_yfieldtmp YField

	num := 1
	for _, j := range cfg {

		LTTest_yfieldtmp.Name = j.Name
		LTTest_yfieldtmp.Description = j.Description
		LTTest_yfieldtmp.Statut = statut
		LTTest_yfieldtmp.Value = JsonINfluxFiledParseFloat(i[0][num])
		num++
		LTTest_yfield = append(LTTest_yfield, LTTest_yfieldtmp)
	}

	return LTTest_yfield

}
