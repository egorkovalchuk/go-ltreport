package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"os"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"
)

var (
	logs           *logger.LogWriter
	loggerOnce     sync.Once
	reportfilename string
	// время формирования отчетов
	timeperiodstart time.Time
	timeperiodend   time.Time
	// Пользовательский период формировани
	EndDateStr   string
	StartDateStr string
	EndDate      time.Time
	StartDate    time.Time
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
		cfg.Confluence.ReportConfluenceProxy = ConflProxy
	}

	if ConfToken != "" {
		cfg.Confluence.ReportConfluenceToken = ConfToken
	}

	if CHUser != "" && CHPass != "" {
		cfg.ClickHouse.User = CHUser
		cfg.ClickHouse.Pass = CHPass
	}

	cfg.ReportPath = reportdata.EnsureTrailingSeparator(cfg.ReportPath)

}

func InitTime() {
	// Формирование имени выхожного файла
	currentTime := time.Now()
	reportfilename = cfg.ReportFilename + currentTime.Format(cfg.ReportMask)

	if hour {
		logs.ProcessInfo("Start group by hour")
		timeperiodstart = reportdata.BeginningOfHour()
		timeperiodend = reportdata.EndOfHour()
		reportfilename = reportfilename + "_" + strconv.Itoa(timeperiodstart.Local().Hour())
	} else if !StartDate.IsZero() && !EndDate.IsZero() {
		logs.ProcessInfo("Start group arbitrary period")
		timeperiodstart = StartDate
		timeperiodend = EndDate
		reportfilename = reportfilename + "_" + "00" + "_" + strconv.Itoa(timeperiodstart.Local().Hour())
	} else {
		logs.ProcessInfo("Start group by day")
		timeperiodstart, timeperiodend = FindTimeWorkTest()
	}
}

func FindTimeWorkTest() (time.Time, time.Time) {
	gc := reportdata.NewInfluxClient(cfg.Jmeter.JmeterInflux, "", reportdata.BeginningOfDay(), reportdata.EndOfDay(), logs, debugm)

	infjson, err := gc.GetDataMean(url.QueryEscape("SELECT max(\"rate\") FROM \"delta\" WHERE" + " " + gc.Timeperiod + " " + "GROUP BY time(15m) fill(0)"))

	if err != nil {
		logs.ProcessWarm("FindTimeWorkTest error")
		logs.ProcessWarm(err)
		logs.ProcessDebug("Start generate report with default time(09:00-19:00)")
		return reportdata.BeginningOfDay(), reportdata.EndOfDay()
	}
	// Проверка наличия данных
	if len(infjson.Results) == 0 || len(infjson.Results[0].Series) == 0 {
		logs.ProcessDebug(len(infjson.Results))
		logs.ProcessDebug(len(infjson.Results[0].Series))
		return reportdata.BeginningOfDay(), reportdata.EndOfDay()
	}

	var min, max int64
	first := true

	for _, i := range infjson.Results[0].Series {
		for _, j := range i.Values {
			if len(j) == 0 {
				continue
			}
			if rate, ok := reportdata.ConvIntefaceFloat64(j[1]); ok && rate > 0 {
				if num, ok := reportdata.ConvIntefaceInt64(j[0]); ok {
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
		logs.ProcessDebug("No valid timestamps found")
	} else {
		logs.ProcessDebug("Time period " + time.Unix(min/1000, 0).Format("02.01.2006 15:04:05") + "-" + time.Unix(max/1000, 0).Format("02.01.2006 15:04:05"))
	}
	return time.Unix(min/1000, 0), time.Unix(max/1000, 0)
}

func RemoveTemp() {

	logs.ProcessInfo("Remove temp files")

	directory, err := os.Getwd()
	if err != nil {
		logs.ProcessError(err)
		return
	}

	readDirectory, err := os.Open(directory + string(filepath.Separator) + "tmp")
	if err != nil {
		logs.ProcessError(err)
		return
	}

	allFiles, err := readDirectory.Readdir(0)
	if err != nil {
		logs.ProcessError(err)
		return
	}

	for f := range allFiles {
		file := allFiles[f]
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".png") {
			os.Remove(directory + string(filepath.Separator) + "tmp" + string(filepath.Separator) + fileName)
		}
	}
}

func createOutputDir(path string) {
	isExists, err := existsDir(path)
	if err != nil {
		logs.ProcessError("Error create temp directory")
	}

	if !isExists {
		_ = os.MkdirAll(path, os.ModePerm)
	}

}

func existsDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func convertEncoding(str, des string) string {

	temp, err := reportdata.ConvertEncoding([]byte(str))
	if err != nil {
		temp = str
		logs.ProcessError(des + ": " + err.Error())
	}
	return temp
}
