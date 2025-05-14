package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"log"
	"os"
	"time"

	"github.com/egorkovalchuk/go-ltreport/reportdata"
)

type LogStruct struct {
	t    string
	text interface{}
}

var (
	LogChannel = make(chan LogStruct)
	// время формирования отчетов
	timeperiod_influx     string
	timeperiod_prometheus string
	timeperiod_clickhouse string
	timeperiod_grafana    string
	timeperiodstart       time.Time
	timeperiodend         time.Time
	// FSM connect
	LoginFSM string
	PassFSM  string
	// Proxy
	// Confluence
	ConflProxy string
	ConfToken  string

	// ClickHouse
	CHUser string
	CHPass string
)

// Запись ошибок из горутин
// можно добавить ротейт по дате + архив в отдельном потоке
func LogWriteForGoRutineStruct(err chan LogStruct) {
	for i := range err {
		datetime := time.Now().Local().Format("2006/01/02 15:04:05")
		log.SetPrefix(datetime + " " + i.t + ": ")
		log.SetFlags(0)
		log.Println(i.text)
		log.SetPrefix("")
		log.SetFlags(log.Ldate | log.Ltime)
	}
}

// Запись в лог при включенном дебаге
// Сделать горутиной?
func ProcessDebug(logtext interface{}) {
	if debugm {
		LogChannel <- LogStruct{"DEBUG", logtext}
	}
}

// Запись в лог ошибок
func ProcessError(logtext interface{}) {
	LogChannel <- LogStruct{"ERROR", logtext}
}

// Запись в лог ошибок cсо множеством переменных
func ProcessErrorAny(logtext ...interface{}) {
	t := ""
	for _, a := range logtext {
		t += fmt.Sprint(a) + " "
	}
	LogChannel <- LogStruct{"ERROR", t}
}

// Запись в лог WARM
func ProcessWarm(logtext interface{}) {
	LogChannel <- LogStruct{"WARM", logtext}
}

// Запись в лог INFO
func ProcessInfo(logtext interface{}) {
	LogChannel <- LogStruct{"INFO", logtext}
}

// Запись в лог Confluence
func ProcessLog(level string, logtext interface{}) {
	LogChannel <- LogStruct{level, logtext}
}

// Нештатное завершение при критичной ошибке
func ProcessPanic(logtext interface{}) {
	fmt.Println(logtext)
	os.Exit(2)
}

// Инициализация переменных
func InitVariables() {}

// Аналог Sleep.
func sleep(d time.Duration) {
	<-time.After(d)
}

// Применение шаблона к датам по произвольному периоду
func DateProcess() bool {
	if StartDateStr != "" && EndDateStr == "" {
		fmt.Println("Error. Missing second parameter \"end\"")
		return false
	}
	if StartDateStr == "" && EndDateStr != "" {
		fmt.Println("Error. Missing second parameter \"start\"")
		return false
	}
	if StartDateStr != "" && EndDateStr != "" {
		if t, err := time.ParseInLocation("2006.01.02 15:04", StartDateStr, time.Now().Location()); err != nil {
			fmt.Println("Error. Use format date for -start \"2023.12.31 15:00\"")
			return false
		} else {
			StartDate = t
		}
		if t, err := time.ParseInLocation("2006.01.02 15:04", EndDateStr, time.Now().Location()); err != nil {
			fmt.Println("Error. Use format date for -end \"2023.12.31 15:00\"")
			return false
		} else {
			EndDate = t
		}
	}
	return true
}

// Чтение конфига
func readconf(cfg *reportdata.Config, confname string) {
	file, err := os.Open(confname)
	if err != nil {
		ProcessPanic(err)
		fmt.Println(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		ProcessPanic(err)
		fmt.Println(err)
	}

	file.Close()
}

func redefinitionconf() {

	// Замена переменных для FSM
	if LoginFSM != "" {
		cfg.ReportIM.LoginFSM = LoginFSM
	}
	if PassFSM != "" {
		cfg.ReportIM.PassFSM = PassFSM
	}

	// Замена переменных Proxy
	if ConflProxy != "" {
		cfg.ReportConfluenceProxy = ConflProxy
	}

	if ConfToken != "" {
		cfg.ReportConfluenceToken = ConfToken
	}

	if CHUser != "" && CHPass != "" {
		cfg.ClickHouse.User = CHUser
		cfg.ClickHouse.Pass = CHPass
	}

}

func InitTime() {
	// Формирование имени выхожного файла
	currentTime := time.Now()
	reportfilename = cfg.ReportFilename + currentTime.Format(cfg.ReportMask)

	if hour {
		ProcessInfo("Start group by hour")
		timeperiodstart = reportdata.BeginningOfHour()
		timeperiodend = reportdata.EndOfHour()
		reportfilename = reportfilename + "_" + strconv.Itoa(timeperiodstart.Local().Hour())
	} else if !StartDate.IsZero() && !EndDate.IsZero() {
		ProcessInfo("Start group arbitrary period")
		timeperiodstart = StartDate
		timeperiodend = EndDate
		reportfilename = reportfilename + "_" + "00" + "_" + strconv.Itoa(timeperiodstart.Local().Hour())
	} else {
		ProcessInfo("Start group by day")
		timeperiodstart, timeperiodend = FindTimeWorkTest()
	}

	timeperiod_grafana = "&from=" + fmt.Sprintf("%d", timeperiodstart.Unix()) + "000&to=" + fmt.Sprintf("%d", timeperiodend.Unix()) + "000"
	timeperiod_influx = ` time >= ` + fmt.Sprintf("%d", timeperiodstart.Unix()) + `000ms AND time <= ` + fmt.Sprintf("%d", timeperiodend.Unix()) + `000ms `
	timeperiod_prometheus = `&start=` + fmt.Sprintf("%d", timeperiodstart.Unix()) + `&end=` + fmt.Sprintf("%d", timeperiodend.Unix())
	timeperiod_clickhouse = " timestamp>=toDateTime('" + timeperiodstart.Format("2006-01-02 15:04:05") + "') and timestamp <=toDateTime('" + timeperiodend.Format("2006-01-02 15:04:05") + "') "
}

func FindTimeWorkTest() (time.Time, time.Time) {
	gc := reportdata.NewInfluxClient(cfg.JmeterInflux, "", ProcessLog, debugm)

	timeperiod := ` time >= ` + fmt.Sprintf("%d", reportdata.BeginningOfDay().Unix()) + `000ms AND time <= ` + fmt.Sprintf("%d", reportdata.EndOfDay().Unix()) + `000ms `
	infjson, err := gc.GetDataMean(url.QueryEscape("SELECT max(\"rate\") FROM \"delta\" WHERE" + " " + timeperiod + " " + "GROUP BY time(15m) fill(0)"))

	if err != nil {
		ProcessWarm("FindTimeWorkTest error")
		ProcessWarm(err)
		ProcessDebug("Start generate report with default time(09:00-19:00)")
		return reportdata.BeginningOfDay(), reportdata.EndOfDay()
	}
	// Проверка наличия данных
	if len(infjson.Results) == 0 || len(infjson.Results[0].Series) == 0 {
		ProcessDebug(len(infjson.Results))
		ProcessDebug(len(infjson.Results[0].Series))
		return reportdata.BeginningOfDay(), reportdata.EndOfDay()
	}

	var min, max int64
	first := true

	for _, i := range infjson.Results[0].Series {
		for _, j := range i.Values {
			if len(j) == 0 {
				continue
			}
			if rate, ok := reportdata.ConvJsonNumFloat64(j[1]); ok && rate > 0 {
				if num, ok := reportdata.ConvJsonNumInt64(j[0]); ok {
					if first {
						min = num
						max = num
						first = false
					} else {
						min = reportdata.MinInt64(min, num)
						max = reportdata.MaxInt64(max, num)
					}
				}
			}
		}
	}
	if first {
		ProcessDebug("No valid timestamps found")
	} else {
		ProcessDebug("Time period " + time.Unix(min/1000, 0).Format("02.01.2006 15:04:05") + "-" + time.Unix(max/1000, 0).Format("02.01.2006 15:04:05"))
	}
	return time.Unix(min/1000, 0), time.Unix(max/1000, 0)
}
