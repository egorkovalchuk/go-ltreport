package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/allure"
	"github.com/egorkovalchuk/go-ltreport/internal/confluence"
	"github.com/egorkovalchuk/go-ltreport/internal/hpsm"
	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/egorkovalchuk/go-ltreport/internal/notify"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"
)

// Power by  Egor Kovalchuk
const (
	//  логи
	logFileName  = "ltreport.log"
	confFileName = "config.json"
	versionutil  = "0.5.0.0"
	a4height     = 297
	a4width      = 210
)

var (
	// Configuration
	cfg       reportdata.Config
	bannotify bool
	// режим работы сервиса(дебаг мод)
	debugm bool
	// запрос помощи
	help bool
	// по часовой отчет за прошедший час
	hour bool
	//  Delete temp files
	rmtmpfile bool
	// запрос версии
	version bool
	// ошибки
	err error
	// Переменная для тестов
	LTTest_dinamic []reportdata.LTTestDinamic
	// Переменная для анализа
	Problems []reportdata.LTError
	// Переменная для сценариев
	LTScenario []reportdata.Scenario
	// Устарело?
	// LTScen_dimanic  map[string]reportdata.ScenarioDinamic
	LTScen_dimanict map[string]map[string]reportdata.ScenarioDinamic
	// Массив графиков и порогов
	LTGrafs []reportdata.LTGrag
	// Пользовательский период формировани
	EndDateStr   string
	StartDateStr string
	EndDate      time.Time
	StartDate    time.Time
	// Аварии
	LTIM hpsm.Content

	// Массив для ClickHouse
	LTClickHouse []reportdata.ClickHouseJson
)

func main() {

	// start program
	//  запуск горутины записи в лог
	loggerOnce.Do(func() {
		logs = logger.NewLogWriter(logFileName, debugm)
		go logs.LogWriteForGoRutineStruct()
	})

	logs.ProcessInfo("- - - - - - - - - - - - - - -")
	logs.ProcessInfo("Start report")

	flag.BoolVar(&debugm, "d", false, "Start debug mode")
	flag.BoolVar(&version, "v", false, "Version")
	var confname string
	flag.StringVar(&confname, "config", confFileName, "start with users config")
	flag.StringVar(&LoginFSM, "fsmlogin", "", "HP Service Manager Login")
	flag.StringVar(&PassFSM, "fsmpass", "", "HP Service Manager Password")
	flag.StringVar(&ConflProxy, "conflproxy", "", "Confluence proxy, use http://user:password@url:port ")
	flag.StringVar(&ConfToken, "ConfToken", "", "Confluence Token ")
	flag.StringVar(&CHUser, "CHUser", "", "ClichHouse User")
	flag.StringVar(&CHPass, "CHPassword", "", "ClichHouse password")
	flag.BoolVar(&help, "h", false, "Use -h for help")
	flag.BoolVar(&hour, "hour", false, "Generate by hourly report")
	flag.BoolVar(&rmtmpfile, "rm", false, "Remove temp files")
	flag.BoolVar(&bannotify, "ban", false, "Ban sending notifications")
	flag.StringVar(&StartDateStr, "start", "", "Start date of report generation in format 2006.01.31 15:00")
	flag.StringVar(&EndDateStr, "end", "", "End date of report generation in format 2006.01.31 15:00")
	flag.Parse()

	readconf(&cfg, confname)
	//  Замена на приоритетный конфиг из командной строки
	redefinitionconf()

	logs.ChangeDebugLevel(debugm)

	if cfg.ReportConfluenceOn && cfg.ReportConfluenceURL == "" {
		fmt.Printf("Confluence URL is required when ReportConfluenceOn=true")
		return
	}

	InitTime()

	//  Получение помощи
	if help {
		reportdata.Helpstart()
		return
	}

	//  получение версии
	if version {
		fmt.Println("Version utils " + versionutil)
		return
	}

	//  Проверка по датам
	if !DateProcess() {
		return
	}

	logs.ProcessDebug("Start with debug mode")
	StartReport()
	RemoveTemp()
	if !bannotify {
		err := notify.SendMessageWithAttach("Report "+reportfilename, "Report for "+timeperiodstart.Format("01\\.02\\.2006 15:04:05")+"\\-"+timeperiodend.Format("01\\.02\\.2006 15:04:05"), cfg.ReportPath+reportfilename+".pdf")
		if err != nil {
			logs.ProcessError(err)
		}
	}
	sleep(2)
}

func StartReport() {

	fmt.Println("Start report")
	logs.ProcessInfo("Report generation")

	ReportPDFInit()

	// Получение инцидентов
	if cfg.ReportOn.ReportIM {
		ReportIM()
	}
	// заведение работ
	// Загрузка данных из графаны
	if cfg.ReportOn.ReportJmeter {
		ReportInflux()
	}
	// Загрузка данных из графаны по сценариям
	if cfg.ReportOn.ReportScenario {
		InfluxJmeterScenario()
	}

	if cfg.ReportOn.ReportDash {
		GrafanaReport()
	}

	// Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		ClickHouseReport()
	}

	// Формирование отчета
	// Обязательно, поэтому не исключаем
	ReportProblemPDF()

	// Включение графиков
	if cfg.ReportOn.ReportDash {
		GrafanaReportPDF()
	}

	// Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		ClickHouseReportPDF()
	}

	if cfg.ReportOn.ReportIM {
		ReportIMPDF()
	}
	// Прогрузка общенй информации
	if cfg.ReportOn.ReportJmeter {
		ReportInfluxPDF()
		ReportProblemScenPDF()
		ReportInfluxScrnPDF()
	}

	ReportEnd()

	if cfg.ReportConfluenceOn {
		logs.ProcessInfo("Start load report on confluence")
		ReportDownload(reportfilename + ".pdf")
	}
	//logs.ProcessDebug(Problems)
	//logs.ProcessDebug(LTGrafs)
}

func ReportInflux() {

	// Получение данных из инфлюкса jmeter
	InfluxErrorJmeter()

	var JMeterTestTh map[string]map[string]reportdata.KeyField
	JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

	// Построение карты порогов
	logs.ProcessInfo("Load map threshold for tests ")
	for _, j := range cfg.Jmeter.JmeterQueryThreshold {
		JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: ""})
		logs.ProcessDebug(j.Name + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	// Формирование стека ошибок
	for _, ii := range LTTest_dinamic {

		logs.ProcessInfo("Load stack error for " + ii.NameTest)

		// Есть ли пороги для данного теста
		if val, ok := JMeterTestTh[ii.NameTest]; ok {
			logs.ProcessDebug(val)

			for _, jj := range ii.Field {
				// Берем имя поля с порогом
				// Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				// Если есть определяем порог
				if valthershold, okk := JMeterTestTh[ii.NameTest][jj.Name]; okk {

					if jj.Value > float64(valthershold.Value) {
						logs.ProcessDebug(ii.NameTest + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						p := reportdata.LTError{Name: ii.NameTest, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value))}
						Problems = append(Problems, p)

					}
				}
			}
		}
		if val, ok := JMeterTestTh["*"]; ok {
			logs.ProcessDebug(val)
			for _, jj := range ii.Field {
				// Берем имя поля с порогом
				// Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				// Если есть определяем порог
				if valthershold, okk := JMeterTestTh["*"][jj.Name]; okk {
					if jj.Value > float64(valthershold.Value) {
						logs.ProcessDebug(ii.NameTest + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						p := reportdata.LTError{Name: ii.NameTest, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value))}
						Problems = append(Problems, p)

					}
				}
			}
		}

	}

}

func InfluxErrorJmeter() {

	logs.ProcessInfo("Load Jmeter delta")
	gc := reportdata.NewInfluxClient(cfg.Jmeter.JmeterInflux, "", logs, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.Jmeter.JmeterQuery + " " + timeperiod_influx + " " + cfg.Jmeter.JmeterQueryGroup))

	if err != nil {
		logs.ProcessError("InfluxErrorJmeter error")
		logs.ProcessError(err)
		return
	}

	for _, i := range infjson.Results[0].Series {

		logs.ProcessInfo("Test " + i.Tags.Suite + " load")

		var LTTest_yfield []reportdata.YField
		var LTTest_dinamictmp reportdata.LTTestDinamic
		var LTTest_yfieldtmp reportdata.YField

		LTTest_dinamictmp.NameTest = i.Tags.Suite

		num := 1
		for _, j := range cfg.Jmeter.JmeterQueryField {

			LTTest_yfieldtmp.Name = j.Name
			LTTest_yfieldtmp.Description = j.Description
			LTTest_yfieldtmp.Value = gc.JsonINfluxFiledParseFloat(i.Values[0][num])
			num++
			LTTest_yfield = append(LTTest_yfield, LTTest_yfieldtmp)
		}

		LTTest_dinamictmp.Field = LTTest_yfield
		LTTest_dinamic = append(LTTest_dinamic, LTTest_dinamictmp)

		logs.ProcessDebug(LTTest_dinamictmp)
	}
}

func InfluxJmeterScenario() {

	logs.ProcessInfo("Jmeter Scenario")
	gc := reportdata.NewInfluxClient(cfg.Jmeter.JmeterInflux, "", logs, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.Jmeter.JmeterQueryScenario + timeperiod_influx + cfg.Jmeter.JmeterQueryScnrGroup))

	if err != nil {
		logs.ProcessError("InfluxJmeterScenario error")
		logs.ProcessError(err)
		return
	}
	var JMeterTestTh map[string]map[string]reportdata.KeyField
	JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

	LTScen_dimanict = make(map[string]map[string]reportdata.ScenarioDinamic)

	// Построение карты порогов
	logs.ProcessInfo("Load map threshold for Scenario ")
	for _, j := range cfg.Jmeter.JmeterQueryScnrThreshold {
		JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name+":"+j.NameThread, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: j.Statut})
		logs.ProcessDebug(j.Name + ":" + j.NameThread + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	for _, i := range infjson.Results[0].Series {
		if i.Tags.Statut != "" && i.Tags.Transaction != "all" {

			logs.ProcessDebug("Load scenario " + i.Tags.Application + ":" + i.Tags.Transaction + " for statut " + i.Tags.Statut)
			// можно так но все равно потом цикл в цикле :(
			LTScenTmpt := LTScen_dimanict[i.Tags.Application][i.Tags.Transaction]
			LTScenTmpt.SetApplication(i.Tags.Application)
			LTScenTmpt.SetThread(i.Tags.Transaction)
			LTScenTmpt.SeField(gc.InfluxJmeterScenarioStatut(i.Values, i.Tags.Statut, cfg.Jmeter.JmeterQueryScnrField))
			LTScen_dimanict = reportdata.AddMapS(LTScen_dimanict, i.Tags.Application, i.Tags.Transaction, LTScenTmpt)
		}

	}

	for _, k := range LTScen_dimanict {
		for _, ii := range k {

			// Есть ли пороги для данного теста
			if val, ok := JMeterTestTh[ii.NameTest+":"+ii.NameThread]; ok {
				logs.ProcessDebug("Load stack error for " + ii.NameTest + ":" + ii.NameThread)
				logs.ProcessDebug(val)

				for _, jj := range ii.Field {
					// Берем имя поля с порогом
					// Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
					// Если есть определяем порог
					if valthershold, okk := JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name]; okk {

						if jj.Value > float64(valthershold.Value) && jj.Statut == JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							p := reportdata.LTError{Name: ii.NameTest + ":" + ii.NameThread, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), Type: "Jmeter"}
							Problems = append(Problems, p)

						}
					}

				}
				// смотрим дефолты
			} else if _, ok := JMeterTestTh["*:*"]; ok {

				for _, jj := range ii.Field {
					// Берем имя поля с порогом
					// Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
					// Если есть определяем порог
					if valthershold, okk := JMeterTestTh["*:*"][jj.Name]; okk {

						if jj.Value > float64(valthershold.Value) && jj.Statut == JMeterTestTh["*:*"][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							p := reportdata.LTError{Name: ii.NameTest + ":" + ii.NameThread, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), Type: "Jmeter"}
							Problems = append(Problems, p)

						}

					}
				}
			} else if _, ok := JMeterTestTh[ii.NameTest+":*"]; ok {

				for _, jj := range ii.Field {
					// Берем имя поля с порогом
					// Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
					// Если есть определяем порог
					if valthershold, okk := JMeterTestTh[ii.NameTest+":*"][jj.Name]; okk {

						if jj.Value > float64(valthershold.Value) && jj.Statut == JMeterTestTh[ii.NameTest+":*"][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							p := reportdata.LTError{Name: ii.NameTest + ":" + ii.NameThread, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), Type: "Jmeter"}
							Problems = append(Problems, p)

						}

					}
				}
			}
		}
	}
}

func GrafanaReport() {

	logs.ProcessInfo("Load grafana metrics")
	alluretmp := allure.NewAllure(cfg.ReportAllure, logs)

	var p reportdata.LTGrag
	for _, i := range cfg.Grafanadash {

		logs.ProcessInfo("Load grafana " + i.Name)
		request := i.Urlimg + timeperiod_grafana
		logs.ProcessDebug("Get image request " + request)

		gc := reportdata.NewGrafanaClient(request, i.AuthHeader, logs, debugm)
		ConType, err := gc.GetImage(i.Name)
		defer gc.Close()

		if err != nil {
			logs.ProcessError("Error generate image")
			logs.ProcessError(err)
		}

		p.Name = i.Name
		p.Threshold = i.Threshold
		p.ContentType = ConType
		p.Size.Height = i.Size.Height
		p.Size.Width = i.Size.Width
		p.UrlDash = i.Urldash + timeperiod_grafana
		LTGrafs = append(LTGrafs, p)

		if i.Query == "" {
			continue
		}

		var percentile float64
		if i.SourceType == 2 {
			//  получение данные из прометеуса
			gcs := reportdata.NewPrometheusClient(i.UrlQuery, i.AuthHeader, logs, debugm)
			percentile, err = gcs.GetThreshold(url.QueryEscape(i.Query+" "+i.UrlQueryGroup) + timeperiod_prometheus)
			defer gcs.Close()
		} else if i.SourceType == 3 {
			gh := reportdata.NewGraphiteClient(i.UrlQuery, i.AuthHeader, logs, debugm)
			percentile, err = gh.Get99thPercentile(i.Query, timeperiodstart, timeperiodend)
			defer gh.Close()
		} else {
			//  получение данные из инфлюкса
			gcs := reportdata.NewInfluxClient(i.UrlQuery, i.AuthHeader, logs, debugm)
			percentile, err = gcs.GetThreshold(url.QueryEscape(i.Query + " AND " + timeperiod_influx + i.UrlQueryGroup))
			defer gcs.Close()
		}
		logs.ProcessDebug("Load threshold: " + fmt.Sprintf("%f", percentile))
		if percentile > float64(i.Threshold) && err == nil {
			ltp := reportdata.LTError{Name: "Grafana: " + i.Name, Threshold: i.Threshold, Description: i.ThDescription + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)) + "", Type: "Grafana"}
			Problems = append(Problems, ltp)

			alluretmp.CreateAllureReport(i.ThDescription, i.Name, fmt.Sprintf("%s[%s start %d]", i.ThDescription, i.Name, timeperiodstart.UnixMilli()), "failed", "finished", timeperiodstart.UnixMilli(), timeperiodend.UnixMilli(), []allure.AllureLabel{{Name: "severity", Value: "critical"}, {Name: "feature", Value: i.ThDescription}}, []allure.AllureParameter{{Name: "Threshold", Value: strconv.Itoa(i.Threshold)}, {Name: "Value", Value: strconv.Itoa(int(percentile))}}, []allure.AllureLink{{Name: "Grafana", URL: p.UrlDash, Type: "requirement"}})
		} else {
			alluretmp.CreateAllureReport(i.ThDescription, i.Name, fmt.Sprintf("%s[%s start %d]", i.ThDescription, i.Name, timeperiodstart.UnixMilli()), "passed", "finished", timeperiodstart.UnixMilli(), timeperiodend.UnixMilli(), []allure.AllureLabel{{Name: "severity", Value: "critical"}, {Name: "feature", Value: i.ThDescription}}, []allure.AllureParameter{{Name: "Threshold", Value: strconv.Itoa(i.Threshold)}, {Name: "Value", Value: strconv.Itoa(int(percentile))}}, []allure.AllureLink{{Name: "Grafana", URL: p.UrlDash, Type: "requirement"}})
		}
	}
	GrafanaTemplateReport()
}

func GrafanaTemplateReport() {
	logs.ProcessInfo("Load grafana template metrics")
	alluretmp := allure.NewAllure(cfg.ReportAllure, logs)

	for _, i := range cfg.GrafanadashTemplate {
		if i.FileList != "" {
			f, err := os.Open(i.FileList)
			if err != nil {
				logs.ProcessErrorAny("Unable to read input file "+i.FileList, err)
				logs.ProcessError("Thread " + i.Name + " not start")
			} else {
				defer f.Close()
				logs.ProcessDebug("Start load " + i.FileList)

				// read csv values using csv.Reader
				csvReader := csv.NewReader(f)
				// Разделитель CSV
				csvReader.Comma = ';'
				csv, err := csvReader.ReadAll()

				if err != nil {
					logs.ProcessError(err)
					continue
				}

				var result []reportdata.DinamicRecord

				headers := strings.Split(i.FilePettern, ";")
				colunmlen := len(headers)

				for j, line := range csv {
					// Проверяем соответствие количества столбцов
					if len(line) != colunmlen {
						logs.ProcessError(fmt.Errorf("Row %d: does not match the number of columns ", j+1))
						continue
					}
					record := make(reportdata.DinamicRecord, colunmlen)

					tmp_dash := i.Urldash
					tmp_query := i.Query
					tmp_image := i.Urlimg
					tmp_name := i.Name
					for jj, head := range headers {
						record[head] = line[jj]
						tmp_dash = strings.Replace(tmp_dash, "{"+head+"}", line[jj], 1)
						tmp_query = strings.Replace(tmp_query, "{"+head+"}", line[jj], 1)
						tmp_image = strings.Replace(tmp_image, "{"+head+"}", line[jj], 1)
						tmp_name = strings.Replace(tmp_name, "{"+head+"}", line[jj], 1)
					}
					result = append(result, record)

					var percentile float64
					if i.SourceType == 2 {
						//  получение данные из прометеуса
						gcs := reportdata.NewPrometheusClient(i.UrlQuery, i.AuthHeader, logs, debugm)
						percentile, err = gcs.GetThreshold(url.QueryEscape(i.Query+" "+i.UrlQueryGroup) + timeperiod_prometheus)
						defer gcs.Close()
					} else if i.SourceType == 3 {
						gh := reportdata.NewGraphiteClient(i.UrlQuery, i.AuthHeader, logs, debugm)
						percentile, err = gh.Get99thPercentile(tmp_query, timeperiodstart, timeperiodend)
						defer gh.Close()
					} else {
						//  получение данные из инфлюкса
						gcs := reportdata.NewInfluxClient(i.UrlQuery, i.AuthHeader, logs, debugm)
						percentile, err = gcs.GetThreshold(url.QueryEscape(i.Query + " AND " + timeperiod_influx + i.UrlQueryGroup))
						defer gcs.Close()
					}
					if percentile > float64(i.Threshold) && err == nil {
						logs.ProcessDebug(tmp_name + " " + strings.Join(line, ", ") + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)))
						p := reportdata.LTError{Name: "Grafana: " + tmp_name + " " + strings.Join(line, ", "), Threshold: i.Threshold, Description: i.ThDescription + " " + strings.Join(line, ", ") + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)) + "", Type: "Grafana"}
						Problems = append(Problems, p)

						gc := reportdata.NewGrafanaClient(tmp_image+timeperiod_grafana, i.AuthHeader, logs, debugm)
						ConType, err := gc.GetImage(tmp_name)
						defer gc.Close()

						if err != nil {
							logs.ProcessError("Error generate image")
							logs.ProcessError(err)
						} else {
							var p reportdata.LTGrag
							p.Name = tmp_name
							p.Threshold = i.Threshold
							p.ContentType = ConType
							p.Size.Height = i.Size.Height
							p.Size.Width = i.Size.Width
							p.UrlDash = tmp_dash + timeperiod_grafana
							LTGrafs = append(LTGrafs, p)
						}
						alluretmp.CreateAllureReport(i.ThDescription, tmp_name, fmt.Sprintf("%s[%s start %d]", i.ThDescription, tmp_name, timeperiodstart.UnixMilli()), "failed", "finished", timeperiodstart.UnixMilli(), timeperiodend.UnixMilli(), []allure.AllureLabel{{Name: "severity", Value: "critical"}, {Name: "feature", Value: i.ThDescription}}, []allure.AllureParameter{{Name: "Threshold", Value: strconv.Itoa(i.Threshold)}, {Name: "Value", Value: strconv.Itoa(int(percentile))}}, []allure.AllureLink{{Name: "Grafana", URL: tmp_dash + timeperiod_grafana, Type: "requirement"}})

					} else if err != nil {
						logs.ProcessError(err)
						alluretmp.CreateAllureReport(i.ThDescription, tmp_name, fmt.Sprintf("%s[%s start %d]", i.ThDescription, tmp_name, timeperiodstart.UnixMilli()), "failed", "finished", timeperiodstart.UnixMilli(), timeperiodend.UnixMilli(), []allure.AllureLabel{{Name: "severity", Value: "critical"}, {Name: "feature", Value: i.ThDescription}}, nil, nil)
					} else {
						alluretmp.CreateAllureReport(i.ThDescription, tmp_name, fmt.Sprintf("%s[%s start %d]", i.ThDescription, tmp_name, timeperiodstart.UnixMilli()), "passed", "finished", timeperiodstart.UnixMilli(), timeperiodend.UnixMilli(), []allure.AllureLabel{{Name: "severity", Value: "critical"}, {Name: "feature", Value: i.ThDescription}}, []allure.AllureParameter{{Name: "Threshold", Value: strconv.Itoa(i.Threshold)}, {Name: "Value", Value: strconv.Itoa(int(percentile))}}, []allure.AllureLink{{Name: "Grafana", URL: tmp_dash + timeperiod_grafana, Type: "requirement"}})
					}
				}
			}
		} else {
			logs.ProcessInfo("Pool not defined for " + i.Name + " template dash")
		}
	}
}

func ClickHouseReport() {
	logs.ProcessInfo("Start load ClickHouse")

	ch := reportdata.NewCHClient("http://"+cfg.ClickHouse.Server+"/?", cfg.ClickHouse.User, cfg.ClickHouse.Pass, logs, debugm)
	for _, i := range cfg.ClickHouse.Query {
		clkhouse, err := ch.GetSql(i.DBname, i.Sql, i.Name, timeperiod_clickhouse)
		if err == nil {
			LTClickHouse = append(LTClickHouse, clkhouse)
		} else {
			logs.ProcessError(err)
		}
	}
	defer ch.Close()
}

func ReportIM() {
	logs.ProcessInfo("Start report FSM")
	fsm := hpsm.NewFSM(cfg.ReportIM.Fsmconnect+cfg.ReportIM.Fsmtu, cfg.ReportIM.Token, cfg.ReportIM.LoginFSM, cfg.ReportIM.PassFSM, logs)
	LTIM = fsm.GetIM()
}

// Загрузка в джиру
func ReportDownload(reportfilename string) {
	// пробрасываем дебаг
	confluence.DebugFlag = debugm
	// Указваем что писать в лог
	confluence.LogFlag = true

	// Инициализация работы с конфленсом
	// перенеммные потом вынески в конфиг
	confl, err := confluence.NewAPI(cfg.ReportConfluenceURL, cfg.ReportConfluenceLogin, cfg.ReportConfluencePass, cfg.ReportConfluenceToken, cfg.ReportConfluenceProxy)
	if err != nil {
		logs.ProcessError("Error connection to confluence")
		logs.ProcessError(err)
		return
	}
	// Получение описание базовой страницы
	logs.ProcessDebug("GetContent")
	JsonCont, err := confl.GetContent(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Expand: []string{"children.page"}}, logs.ProcessLog)
	if err != nil {
		logs.ProcessError(err)
		logs.ProcessError(JsonCont)
		return
	}
	logs.ProcessDebug("GetContentChildPage")
	JsonConC, err := confl.GetContentChildPage(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Limit: 250, Expand: []string{"children.page"}}, logs.ProcessLog)
	if err != nil {
		logs.ProcessError(err)
		logs.ProcessError(JsonConC)
		return
	}

	currentTime := time.Now()
	reportname := "Report" + currentTime.Format("20060102")

	var IdChild string

	// поиск по детям GetContentChildPage
	for _, i := range JsonConC.Results {
		if reportname == i.Title {
			IdChild = i.ID
			logs.ProcessDebug(i.Title + ", id=" + IdChild)
		} else {
			IdChild = ""
		}
	}

	if IdChild == "" {
		logs.ProcessDebug("Create child page " + reportname)
		// формирование тела для создания
		data := confluence.ConflCreateType{
			Type:  "page",
			Title: reportname,
			Ancestors: []confluence.Ancestor{
				{
					ID: JsonCont.ID,
				},
			},
			Body: confluence.Body{
				Storage: confluence.Storage{
					Value:          "Load testing report for " + currentTime.Format("02.01.2006") + "<br/> See Attachments",
					Representation: "storage",
				},
			},
			Version: &confluence.Version{
				Number: 1,
			},
			Space: confluence.Space{
				Key: cfg.ReportConfluenceSpace,
			},
		}

		JsonContC, err := confl.CreateContent(&data, logs.ProcessLog)
		if err != nil {

			logs.ProcessError(err)
			return
		}
		IdChild = JsonContC.ID

		file, err := os.OpenFile(cfg.ReportPath+reportfilename, os.O_RDONLY, 0666)
		if err != nil {
			logs.ProcessError(err)
		}

		logs.ProcessDebug("Upload current attachments " + reportfilename)
		arsp, err := confl.UploadAttachment(IdChild, reportfilename, file, logs.ProcessLog)
		if err != nil {
			logs.ProcessError(err)
			logs.ProcessDebug(arsp)
		}

		defer file.Close()

	} else {
		logs.ProcessDebug("Load current attachments")
		arsp, err := confl.GetAttachments(IdChild, logs.ProcessLog)
		if err != nil {
			logs.ProcessError(err)
			logs.ProcessDebug(arsp)
		}

		chck := false
		attachid := "0"
		for _, j := range arsp.Results {
			if j.Title == reportfilename {
				attachid = j.ID
				chck = true
			}
		}

		file, err := os.OpenFile(cfg.ReportPath+reportfilename, os.O_RDONLY, 0666)
		if err != nil {
			logs.ProcessInfo(err)
		}

		if chck {
			logs.ProcessDebug("Update current attachments " + reportfilename)
			arsp, err := confl.UpdateAttachment(IdChild, reportfilename, attachid, file, logs.ProcessLog)
			if err != nil {
				logs.ProcessInfo(err)
				logs.ProcessDebug(arsp)
			}
		} else {
			logs.ProcessDebug("Upload current attachments " + reportfilename)
			arsp, err := confl.UploadAttachment(IdChild, reportfilename, file, logs.ProcessLog)
			if err != nil {
				logs.ProcessInfo(err)
				logs.ProcessDebug(arsp)
			}
		}

		defer file.Close()

	}

}

func RemoveTemp() {

	logs.ProcessInfo("Remove temp files")

	directory, _ := os.Getwd()
	readDirectory, _ := os.Open(directory)
	allFiles, _ := readDirectory.Readdir(0)

	for f := range allFiles {
		file := allFiles[f]
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".png") {
			os.Remove(fileName)
		}
	}
}
