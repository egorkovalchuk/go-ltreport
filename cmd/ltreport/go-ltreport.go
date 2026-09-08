package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/allure"
	"github.com/egorkovalchuk/go-ltreport/internal/hpsm"
	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/egorkovalchuk/go-ltreport/internal/notify"
	"github.com/egorkovalchuk/go-ltreport/internal/pdf"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"
	"github.com/egorkovalchuk/go-ltreport/internal/source/clickhouse"
	"github.com/egorkovalchuk/go-ltreport/internal/source/grafana"
	"github.com/egorkovalchuk/go-ltreport/internal/source/graphite"
	"github.com/egorkovalchuk/go-ltreport/internal/source/influx"
	"github.com/egorkovalchuk/go-ltreport/internal/source/prometheus"
	"github.com/google/uuid"
)

// Power by  Egor Kovalchuk
const (
	//  логи
	logFileName  = "ltreport.log"
	confFileName = "config.json"
	versionutil  = "0.6.0.2"
)

var (
	// Configuration
	cfg       reportdata.Config
	bannotify bool
	// режим работы сервиса(дебаг мод)
	debugm bool
	// по часовой отчет за прошедший час
	hour bool
	//  Delete temp files
	rmtmpfile bool
	// Переменная для тестов
	LTTest_dinamic influx.LTTestDinamics
	// Переменная для анализа
	Problems reportdata.LTErrors
	// Сценарии
	LTScen_dimanict map[string]map[string]influx.ScenarioDinamic
	// Массив графиков и порогов
	LTGrafs reportdata.LTGrags
	// Аварии
	LTIM hpsm.Content
	// allure
	alluretmp *allure.Allure
	// Массив для ClickHouse
	LTClickHouse []clickhouse.ClickHouseJson
)

func main() {

	// start program
	var version, help bool
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

	// запуск горутины записи в лог
	loggerOnce.Do(func() {
		logs = logger.NewLogWriter(logFileName, debugm)
		go logs.LogWriteForGoRutineStruct()
	})

	logs.ProcessInfo("- - - - - - - - - - - - - - -")
	logs.ProcessInfo("Start report")

	err := cfg.Readconf(confname)
	if err != nil {
		logs.ProcessPanic(err)
	}
	//  Замена на приоритетный конфиг из командной строки
	redefinitionconf()

	logs.ChangeDebugLevel(debugm)

	if cfg.Confluence.ReportConfluenceOn && cfg.Confluence.ReportConfluenceURL == "" {
		fmt.Printf("Confluence URL is required when ReportConfluenceOn=true")
		return
	}

	InitTime()
	createOutputDir("tmp/")

	//  Проверка по датам
	if !DateProcess() {
		return
	}

	logs.ProcessDebug("Start with debug mode")
	StartReport()
	RemoveTemp()
	if !bannotify {
		logs.ProcessInfo("Start notification")
		err := notify.SendMessageWithAttach(reportfilename, "Report for "+timeperiodstart.Format("01.02.2006 15:04:05")+"-"+timeperiodend.Format("15:04:05"), cfg.ReportPath+reportfilename+".pdf")
		if err != nil {
			logs.ProcessError(err)
		}
	}
	sleep(2)
}

func StartReport() {

	fmt.Println("Start report")
	logs.ProcessInfo("Report generation")

	alluretmp = allure.NewAllure(cfg.Allure.ReportAllure, cfg.Allure.Token, cfg.Allure.ReportEndPoint, cfg.Allure.ReportPath)
	go func(ec <-chan *allure.LogStruct) {
		for err := range ec {
			t := err
			logs.ProcessLog(t.T, t.Text)
		}
	}(alluretmp.GetError())

	pdf := pdf.NewPDFWhriter(reportfilename, cfg.ReportPath, timeperiodstart, timeperiodend, cfg.Confluence, logs)
	pdf.ReportPDFInit()

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
		GrafanaTemplateReport()
	}

	// Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		ClickHouseReport()
	}

	// Формирование отчета
	// Обязательно, поэтому не исключаем
	pdf.ReportProblemPDF(Problems)

	// Включение графиков
	if cfg.ReportOn.ReportDash {
		pdf.GrafanaReportPDF(LTGrafs)
	}

	// Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		pdf.ClickHouseReportPDF(LTClickHouse)
	}

	if cfg.ReportOn.ReportIM {
		pdf.ReportIMPDF(LTIM)
	}
	// Прогрузка общенй информации
	if cfg.ReportOn.ReportJmeter {
		pdf.ReportInfluxPDF(LTTest_dinamic)
		pdf.ReportProblemScenPDF(Problems)
		pdf.ReportInfluxScrnPDF(LTScen_dimanict)
	}

	pdf.ReportEnd()

	allurelink := ""
	if cfg.Confluence.ReportConfluenceOn {
		logs.ProcessInfo("Start load report on confluence")
		allurelink = pdf.ReportDownload()
	}

	err := alluretmp.Finish("allure-report-"+time.Now().Format(cfg.ReportMask)+".zip", timeperiodstart, timeperiodend, cfg.Confluence.ReportConfluenceURL+allurelink)
	if err != nil {
		logs.ProcessError(err)
	}
	//logs.ProcessDebug(Problems)
	//logs.ProcessDebug(LTGrafs)
}

func ReportInflux() {

	// Получение данных из инфлюкса jmeter
	InfluxErrorJmeter()

	var JMeterTestTh map[string]map[string]influx.KeyField
	JMeterTestTh = make(map[string]map[string]influx.KeyField)

	// Построение карты порогов
	logs.ProcessInfo("Load map threshold for tests ")
	for _, j := range cfg.Jmeter.JmeterQueryThreshold {
		JMeterTestTh = influx.AddMap(JMeterTestTh, j.Name, j.ErrorField, influx.KeyField{Value: j.Threshold, Description: j.Description, Statut: ""})
		logs.ProcessDebug(j.Name + " threshold " + fmt.Sprint(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
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

					if jj.Value > valthershold.Value {
						logs.ProcessDebug(ii.NameTest + " threshold: " + fmt.Sprint(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						Problems.AddError(ii.NameTest, valthershold.Value, fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value)), 0, "Jmeter", "Jmeter")
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
					if jj.Value > valthershold.Value {
						logs.ProcessDebug(ii.NameTest + " threshold: " + fmt.Sprint(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						Problems.AddError(ii.NameTest, valthershold.Value, fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value)), 0, "Jmeter", "Jmeter")
					}
				}
			}
		}
	}
}

func InfluxErrorJmeter() {

	logs.ProcessInfo("Load Jmeter delta")
	gc := influx.NewInfluxClient(cfg.Jmeter.JmeterInflux, "", timeperiodstart, timeperiodend, logs, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.Jmeter.JmeterQuery + " " + gc.Timeperiod + " " + cfg.Jmeter.JmeterQueryGroup))

	if err != nil {
		logs.ProcessError("InfluxErrorJmeter error")
		logs.ProcessError(err)
		return
	}

	for _, i := range infjson.Results[0].Series {

		logs.ProcessInfo("Test " + i.Tags.Suite + " load")

		var LTTest_yfield []influx.YField
		var LTTest_dinamictmp influx.LTTestDinamic
		var LTTest_yfieldtmp influx.YField

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
	gc := influx.NewInfluxClient(cfg.Jmeter.JmeterInflux, "", timeperiodstart, timeperiodend, logs, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.Jmeter.JmeterQueryScenario + gc.Timeperiod + cfg.Jmeter.JmeterQueryScnrGroup))

	if err != nil {
		logs.ProcessError("InfluxJmeterScenario error")
		logs.ProcessError(err)
		return
	}
	var JMeterTestTh map[string]map[string]influx.KeyField
	JMeterTestTh = make(map[string]map[string]influx.KeyField)

	LTScen_dimanict = make(map[string]map[string]influx.ScenarioDinamic)

	// Построение карты порогов
	logs.ProcessInfo("Load map threshold for Scenario ")
	for _, j := range cfg.Jmeter.JmeterQueryScnrThreshold {
		JMeterTestTh = influx.AddMap(JMeterTestTh, j.Name+":"+j.NameThread, j.ErrorField, influx.KeyField{Value: j.Threshold, Description: j.Description, Statut: j.Statut})
		logs.ProcessDebug(j.Name + ":" + j.NameThread + " threshold " + fmt.Sprint(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	for _, i := range infjson.Results[0].Series {
		if i.Tags.Statut != "" && i.Tags.Transaction != "all" {

			logs.ProcessDebug("Load scenario " + i.Tags.Application + ":" + i.Tags.Transaction + " for statut " + i.Tags.Statut)
			// можно так но все равно потом цикл в цикле :(
			LTScenTmpt := LTScen_dimanict[i.Tags.Application][i.Tags.Transaction]
			LTScenTmpt.SetApplication(i.Tags.Application)
			LTScenTmpt.SetThread(i.Tags.Transaction)
			LTScenTmpt.SeField(gc.InfluxJmeterScenarioStatut(i.Values, i.Tags.Statut, cfg.Jmeter.JmeterQueryScnrField))
			LTScen_dimanict = influx.AddMapS(LTScen_dimanict, i.Tags.Application, i.Tags.Transaction, LTScenTmpt)
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

						if jj.Value > valthershold.Value && jj.Statut == JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + fmt.Sprint(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							Problems.AddError(ii.NameTest+":"+ii.NameThread, valthershold.Value, fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), 0, ii.NameTest, "Jmeter")
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

						if jj.Value > valthershold.Value && jj.Statut == JMeterTestTh["*:*"][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + fmt.Sprint(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							Problems.AddError(ii.NameTest+":"+ii.NameThread, valthershold.Value, fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), 0, ii.NameTest, "Jmeter")
						}
					}
				}
			} else if _, ok := JMeterTestTh[ii.NameTest+":*"]; ok {

				for _, jj := range ii.Field {
					// Берем имя поля с порогом
					// Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
					// Если есть определяем порог
					if valthershold, okk := JMeterTestTh[ii.NameTest+":*"][jj.Name]; okk {

						if jj.Value > valthershold.Value && jj.Statut == JMeterTestTh[ii.NameTest+":*"][jj.Name].Statut {
							logs.ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + fmt.Sprint(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
							Problems.AddError(ii.NameTest+":"+ii.NameThread, valthershold.Value, fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), 0, ii.NameTest, "Jmeter")
						}
					}
				}
			}
		}
	}
}

func GrafanaReport() {

	logs.ProcessInfo("Load grafana metrics")

	for _, i := range cfg.Grafanadash {

		logs.ProcessInfo("Load grafana " + i.Name)

		// получение картинки
		gc := grafana.NewGrafanaClient(i.Urlimg, i.AuthHeader, timeperiodstart, timeperiodend, logs, debugm)
		logs.ProcessDebug("Get image request " + i.Urlimg + gc.Timeperiod)
		ConType, err := gc.GetImage("tmp/", i.Name)
		defer gc.Close()

		if err != nil {
			logs.ProcessError("Error generate image")
			logs.ProcessError(err)
		}

		var p reportdata.LTGrag
		p.Name = convertEncoding(i.Name, "Error convert for dash"+i.Name)
		p.Threshold = i.Threshold
		p.ContentType = ConType
		p.Size.Height = i.Size.Height
		p.Size.Width = i.Size.Width
		p.UrlDash = i.Urldash + gc.Timeperiod
		LTGrafs = append(LTGrafs, p)

		i.ThDescription = convertEncoding(i.ThDescription, "Error convert for dash"+i.Name)

		if i.Query == "" && i.AlertID != nil {
			logs.ProcessWarm("Thresholds are not set for " + i.Name)
			continue
		}

		// подготовка записи в аллюр
		rstallure := allure.AllureResult{
			UUID:        uuid.New().String(),
			Name:        i.Name,
			FullName:    i.Name,
			HistoryID:   fmt.Sprintf("%s[%s start %d]", i.Name, i.Name, timeperiodstart.UnixMilli()),
			Start:       timeperiodstart.UnixMilli(),
			Stop:        timeperiodend.UnixMilli(),
			Description: i.ThDescription,
		}

		// Конвертирцем в UTF8
		templabel := convertEncoding(i.Tag, "Error convert for dash"+i.Name)
		rstallure.ArrayToLabelRoot(templabel, "product")
		rstallure.ArrayToLabelRoot("severity:critical;feature:"+i.ThDescription, "product")
		rstallure.AddLink("Grafana", i.Urldash+gc.Timeperiod, "requirement")
		rstallure.AddStep("Init request", "passed", timeperiodstart, timeperiodend)
		err = rstallure.AddAttach(i.Name, allure.PNG, "tmp/"+i.Name+".png", nil)
		if err != nil {
			logs.ProcessError(err)
		}
		// по умолчанию пройден
		rstallure.FilishedPassed()
		// подготовка записи в аллюр

		if i.Query != "" {
			// Получаем данные
			percentile, err := getThreshold(i.SourceType, i.UrlQuery, i.AuthHeader, i.Query, i.UrlQueryGroup)
			statusstep := ""
			if err != nil {
				logs.ProcessError(err)
				rstallure.FilishedBroken()
				rstallure.AddAttach("error.txt", allure.Text, "", []byte(fmt.Sprintf("Error load data - sql: %s\n Error: %s", i.Query, err.Error())))
				statusstep = "broken"
			} else {
				statusstep = "passed"
			}

			// Сохраняем данные в аллюр и проверяем сработку метрики
			rstallure.AddStepWithParam("Checked threshold", statusstep, timeperiodstart, timeperiodend, rstallure.ArrayToParam(fmt.Sprintf("Threshold:%f;Value:%f", i.Threshold, percentile)))
			rstallure.ArrayToParamRoot(fmt.Sprintf("Threshold:%f;Value:%f", i.Threshold, percentile))

			logs.ProcessDebug(fmt.Sprintf("Load threshold: %f; value:%f", i.Threshold, percentile))
			if ((percentile > i.Threshold && i.Comparison == ">") || (percentile < i.Threshold && i.Comparison == "<")) && err == nil {
				Problems.AddError(i.Name, i.Threshold, i.ThDescription, percentile, "Grafana", rstallure.GetValueLabel("product"))
				rstallure.FilishedFailed()
			}
		}
		if i.AlertID != nil {
			logs.ProcessDebug("Alerts " + i.Name)
			alert, txt, err := gc.GetHistAlerts(i.AlertID)
			if err != nil {
				logs.ProcessError(err)
				rstallure.AddStep("Checked alerts", "failed", timeperiodstart, timeperiodend)
			} else {
				rstallure.AddStep("Checked alerts", "passed", timeperiodstart, timeperiodend)
				if alert {
					Problems.AddError("Grafana alert: "+i.Name, 0, txt, 0, "Alerts", rstallure.GetValueLabel("product"))
					logs.ProcessInfo(fmt.Sprintf("The dashboard %s alert has been triggered.", i.Name))
					rstallure.AddAttach("alert.txt", allure.Text, "", []byte(txt))
					rstallure.FilishedFailed()
				}
			}
		}
		// сохраняем результат
		alluretmp.CreateAllureReport(rstallure)
	}
}

func GrafanaTemplateReport() {
	logs.ProcessInfo("Load grafana template metrics")
	var wg sync.WaitGroup
	for _, i := range cfg.GrafanadashTemplate {
		wg.Add(1)
		go func(i reportdata.GrafanadashTemplateStruct) {
			defer wg.Done()
			logs.ProcessDebug("Start thread grafana template")
			if i.FileList == "" {
				logs.ProcessInfo("Pool not defined for " + i.Name + " template dash")
				return
				//continue
			}
			f, err := os.Open(i.FileList)
			if err != nil {
				logs.ProcessErrorAny("Unable to read input file "+i.FileList, err)
				logs.ProcessError("Thread " + i.Name + " not start")
				return
				//continue
			}
			defer f.Close()
			logs.ProcessDebug("Start load " + i.FileList)

			// read csv values using csv.Reader
			csvReader := csv.NewReader(f)
			// Разделитель CSV
			csvReader.Comma = ';'
			csv, err := csvReader.ReadAll()

			if err != nil {
				logs.ProcessError(err)
				return
				//continue
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
				tmp_name := convertEncoding(i.Name, "Error convert for dash"+tmp_dash)

				for jj, head := range headers {
					record[head] = line[jj]
					tmp_dash = strings.Replace(tmp_dash, "{"+head+"}", line[jj], 1)
					tmp_query = strings.Replace(tmp_query, "{"+head+"}", line[jj], 1)
					tmp_image = strings.Replace(tmp_image, "{"+head+"}", line[jj], 1)
					tmp_name = strings.Replace(tmp_name, "{"+head+"}", line[jj], 1)
				}
				result = append(result, record)
				i.ThDescription = convertEncoding(i.ThDescription, "Error convert for dash"+tmp_dash)

				// подготовка записи в аллюр
				rstallure := allure.AllureResult{
					UUID:        uuid.New().String(),
					Name:        tmp_name,
					FullName:    tmp_name,
					HistoryID:   fmt.Sprintf("%s[%s start %d]", i.Name, tmp_name, timeperiodstart.UnixMilli()),
					Start:       timeperiodstart.UnixMilli(),
					Stop:        timeperiodend.UnixMilli(),
					Description: i.ThDescription,
				}

				// Конвертирцем в UTF8
				templabel := convertEncoding(line[1], "Error convert for dash"+tmp_dash)
				rstallure.ArrayToLabelRoot(templabel, "product")
				rstallure.ArrayToLabelRoot(convertEncoding(i.Tag, "Error convert for dash"+tmp_dash), "tag")
				rstallure.ArrayToLabelRoot("severity:critical;feature:"+i.ThDescription, "product")
				rstallure.AddStep("Init request", "passed", timeperiodstart, timeperiodend)
				// подготовка записи в аллюр

				percentile, err := getThreshold(i.SourceType, i.UrlQuery, i.AuthHeader, tmp_query, i.UrlQueryGroup)
				statusstep := ""
				if err != nil {
					statusstep = "broken"
					rstallure.AddAttach("error.txt", allure.Text, "", []byte(fmt.Sprintf("Error load data - sql: %s\n Error: %s", tmp_query, err.Error())))
				} else {
					statusstep = "passed"
				}

				// Сохраняем данные в аллюр и проверяем сработку метрики
				rstallure.AddStepWithParam("Checked threshold", statusstep, timeperiodstart, timeperiodend, rstallure.ArrayToParam(fmt.Sprintf("Threshold:%f;Value:%f", i.Threshold, percentile)))
				rstallure.ArrayToParamRoot(fmt.Sprintf("Threshold:%f;Value:%f", i.Threshold, percentile))
				// Создаем клиент
				gc := grafana.NewGrafanaClient(tmp_image, i.AuthHeader, timeperiodstart, timeperiodend, logs, debugm)
				defer gc.Close()
				rstallure.AddLink("Grafana", tmp_dash+gc.Timeperiod, "requirement")

				if percentile > i.Threshold && err == nil {
					logs.ProcessInfo(tmp_name + " " + strings.Join(line, ", ") + ": Threshold " + fmt.Sprint(i.Threshold) + " - current " + strconv.Itoa(int(percentile)))
					Problems.AddError(tmp_name, i.Threshold, i.ThDescription, percentile, "Grafana", line[1])

					ConType, erri := gc.GetImage("tmp/", tmp_name)
					if erri != nil {
						logs.ProcessError("Error generate image")
						logs.ProcessError(erri)
					} else {
						var p reportdata.LTGrag
						p.Name = tmp_name
						p.Threshold = i.Threshold
						p.ContentType = ConType
						p.Size.Height = i.Size.Height
						p.Size.Width = i.Size.Width
						p.UrlDash = tmp_dash + gc.Timeperiod
						LTGrafs = append(LTGrafs, p)
						// Добавляем картинку в аллюр
						erra := rstallure.AddAttach(tmp_name, allure.PNG, "tmp/"+tmp_name+".png", nil)
						if erra != nil {
							logs.ProcessError(err)
						}
					}

					rstallure.FilishedFailed()
				} else if err != nil {
					logs.ProcessError(err)
					rstallure.FilishedBroken()
				} else {
					rstallure.FilishedPassed()
				}
				// сохраняем результат
				alluretmp.CreateAllureReport(rstallure)
			}
		}(i)
	}
	wg.Wait()
}

func ClickHouseReport() {
	logs.ProcessInfo("Start load ClickHouse")

	ch := clickhouse.NewCHClient("http://"+cfg.ClickHouse.Server+"/?", cfg.ClickHouse.User, cfg.ClickHouse.Pass, timeperiodstart, timeperiodend, logs, debugm)
	for _, i := range cfg.ClickHouse.Query {
		clkhouse, err := ch.GetSql(i.DBname, i.Sql, i.Name)
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
	period := fmt.Sprintf("PeriodStart=%s&PeriodEnd=%s", timeperiodstart.Format("02-01-2006T15:04:05"), timeperiodend.Format("02-01-2006T15:04:05"))
	fsm := hpsm.NewFSM(cfg.ReportIM.Fsmconnect, cfg.ReportIM.Fsmtu, cfg.ReportIM.Token, cfg.ReportIM.LoginFSM, cfg.ReportIM.PassFSM, period, logs)
	LTIM = fsm.GetIM()
}

func getThreshold(SourceType int, UrlQuery, AuthHeader, Query, UrlQueryGroup string) (float64, error) {
	var percentile float64
	var err error
	switch SourceType {
	case 2:
		//  получение данные из прометеуса
		gcs := prometheus.NewPrometheusClient(UrlQuery, AuthHeader, timeperiodstart, timeperiodend, logs, debugm)
		percentile, err = gcs.GetThreshold(Query + " " + UrlQueryGroup)
		defer gcs.Close()
	case 3:
		gh := graphite.NewGraphiteClient(UrlQuery, AuthHeader, logs, debugm)
		percentile, err = gh.Get99thPercentile(Query, timeperiodstart, timeperiodend)
		defer gh.Close()
	default:
		//  получение данные из инфлюкса
		gcs := influx.NewInfluxClient(UrlQuery, AuthHeader, timeperiodstart, timeperiodend, logs, debugm)
		percentile, err = gcs.GetThreshold(Query + " AND " + gcs.Timeperiod + UrlQueryGroup)
		defer gcs.Close()
	}
	return percentile, err
}
