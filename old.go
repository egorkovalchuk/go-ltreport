package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/egorkovalchuk/go-ltreport/reportdata"
)

//Устарело смотри InfluxJmeterScenario()
//Работа с динамикой в InfluxJmeterScenario()
func InfluxJmeterScenarioOld() {

	request := cfg.JmeterInflux + url.QueryEscape(cfg.JmeterQueryScenario+timeperiod+cfg.JmeterQueryScnrGroup)

	ProcessDebug("JmeterScenario")
	ProcessDebug(request)

	resp, err := http.Get(request)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		log.Println("HTTP Status is in the 2xx range " + request)
	} else {
		log.Println("HTTP Status error " + strconv.Itoa(resp.StatusCode) + " " + request)
	}

	infjson, _ := reportdata.JsonINfluxParse(resp)

	resp.Body.Close()

	for _, i := range infjson.Results[0].Series {

		var percen float64
		var maxl float64

		percen = reportdata.JsonINfluxFiledParseFloat(i.Values[0][1])
		maxl = reportdata.JsonINfluxFiledParseFloat(i.Values[0][2])

		ProcessDebug("Scenario " + i.Tags.Transaction)
		ProcessDebug("percentile is " + strconv.Itoa(int(percen)) + " ms")
		ProcessDebug("max latency " + strconv.Itoa(int(maxl)) + " ms")

		request_child := cfg.JmeterInflux + url.QueryEscape(`SELECT mean("count") / 5 FROM "details" WHERE "transaction" ='`+i.Tags.Transaction+`' AND "statut" = 'ok' AND time >= now() - 1d GROUP BY time(1d), "transaction" fill(null) ORDER BY time DESC`)

		ProcessDebug("Load average request rate")
		ProcessDebug(request_child)

		resp, err := http.Get(request_child)
		if err != nil {
			log.Println(err)
		}

		var avgcount float64

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {

			log.Println("HTTP Status is in the 2xx range " + request)
			infjson_child, err := reportdata.JsonINfluxParse(resp)

			if err == nil {

				avgcount = reportdata.JsonINfluxFiledParseFloat(infjson_child.Results[0].Series[0].Values[0][1])

			}
		} else {
			log.Println("HTTP Status error " + strconv.Itoa(resp.StatusCode) + " " + request)
		}

		resp.Body.Close()

		request_child = cfg.JmeterInflux + url.QueryEscape(`SELECT mean("count") / 5 FROM "details" WHERE "transaction" ='`+i.Tags.Transaction+`' AND "statut" = 'ko' AND time >= now() - 1d GROUP BY time(1d), "transaction" fill(null) ORDER BY time DESC`)
		ProcessDebug("Load average request rate")
		ProcessDebug(request_child)

		resp, err = http.Get(request_child)
		if err != nil {
			log.Println(err)
		}

		var avgcountr float64

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			log.Println("HTTP Status is in the 2xx range " + request)

			infjson_child, err := reportdata.JsonINfluxParse(resp)

			if err == nil {

				avgcountr = reportdata.JsonINfluxFiledParseFloat(infjson_child.Results[0].Series[0].Values[0][1])

			}
		} else {
			log.Println("HTTP Status error " + strconv.Itoa(resp.StatusCode) + " " + request)
		}

		resp.Body.Close()

		p := reportdata.Scenario{Tags: i.Tags.Transaction,
			Percentile: percen,
			Maxlatency: maxl,
			Rate:       avgcount,
			RateError:  avgcountr}
		LTScenario = append(LTScenario, p)

	}

}
