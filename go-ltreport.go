package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-ltreport/confluence"
	"github.com/egorkovalchuk/go-ltreport/reportdata"
)

// Power by  Egor Kovalchuk
const (
	//  логи
	logFileName  = "ltreport.log"
	confFileName = "config.json"
	versionutil  = "0.4.0.0"
	a4height     = 297
	a4width      = 210
)

var (
	// Configuration
	cfg reportdata.Config
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

	// режим работы сервиса(дебаг мод)
	debugm bool
	// Запись в лог
	filer *os.File
	// запрос помощи
	help bool
	// по часовой отчет за прошедший час
	hour bool
	//  Delete temp files
	rmtmpfile bool
	// ошибки
	err error
	// запрос версии
	version bool
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

	// Массив для ClickHouse
	LTClickHouse []reportdata.ClickHouseJson
)

func main() {

	// start program
	filer, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(filer)

	//  запуск горутины записи в лог
	go LogWriteForGoRutineStruct(LogChannel)

	ProcessInfo("- - - - - - - - - - - - - - -")
	ProcessInfo("Start report")

	flag.BoolVar(&debugm, "d", false, "Start debug mode")
	flag.BoolVar(&version, "v", false, "Version")
	var confname string
	flag.StringVar(&confname, "c", confFileName, "start with users config")
	flag.StringVar(&LoginFSM, "fsmlogin", "", "HP Service Manager Login")
	flag.StringVar(&PassFSM, "fsmpass", "", "HP Service Manager Password")
	flag.StringVar(&ConflProxy, "conflproxy", "", "Confluence proxy, use http://user:password@url:port ")
	flag.StringVar(&ConfToken, "ConfToken", "", "Confluence Token ")
	flag.StringVar(&CHUser, "CH User", "", "ClichHouse User")
	flag.StringVar(&CHPass, "CH Password", "", "ClichHouse password")
	flag.BoolVar(&help, "h", false, "Use -h for help")
	flag.BoolVar(&hour, "hour", false, "Generate by hourly report")
	flag.BoolVar(&rmtmpfile, "rm", false, "Remove temp files")
	flag.StringVar(&StartDateStr, "start", "", "Start date of report generation in format 2006.01.31 15:00")
	flag.StringVar(&EndDateStr, "end", "", "End date of report generation in format 2006.01.31 15:00")
	flag.Parse()

	readconf(&cfg, confname)
	//  Замена на приоритетный конфиг из командной строки
	redefinitionconf()

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

	ProcessDebug("Start with debug mode")
	StartReport()
	RemoveTemp()
}

func StartReport() {

	fmt.Println("Start report")
	ProcessInfo("Report generation")

	ReportInit()

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

	// Прогрузка общенй информации
	if cfg.ReportOn.ReportJmeter {
		ReportInfluxPDF()
		ReportProblemScenPDF()
		ReportInfluxScrnPDF()
	}

	ReportEnd()

	if cfg.ReportConfluenceOn {
		ReportDownload(reportfilename + ".pdf")
	}

}

func ReportInflux() {

	// Получение данных из инфлюкса jmeter
	InfluxErrorJmeter()

	var JMeterTestTh map[string]map[string]reportdata.KeyField
	JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

	// Построение карты порогов
	ProcessInfo("Load map threshold for tests ")
	for _, j := range cfg.JmeterQueryThreshold {
		JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: ""})
		ProcessDebug(j.Name + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	// Формирование стека ошибок
	for _, ii := range LTTest_dinamic {

		ProcessInfo("Load stack error for " + ii.NameTest)

		// Есть ли пороги для данного теста
		if val, ok := JMeterTestTh[ii.NameTest]; ok {
			ProcessDebug(val)

			for _, jj := range ii.Field {
				// Берем имя поля с порогом
				// Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				// Если есть определяем порог
				if valthershold, okk := JMeterTestTh[ii.NameTest][jj.Name]; okk {

					if jj.Value > float64(valthershold.Value) {
						ProcessDebug(ii.NameTest + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						p := reportdata.LTError{Name: ii.NameTest, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value))}
						Problems = append(Problems, p)

					}
				}
			}
		}
		if val, ok := JMeterTestTh["*"]; ok {
			ProcessDebug(val)
			for _, jj := range ii.Field {
				// Берем имя поля с порогом
				// Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				// Если есть определяем порог
				if valthershold, okk := JMeterTestTh["*"][jj.Name]; okk {
					if jj.Value > float64(valthershold.Value) {
						ProcessDebug(ii.NameTest + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
						p := reportdata.LTError{Name: ii.NameTest, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest, int(jj.Value))}
						Problems = append(Problems, p)

					}
				}
			}
		}

	}

}

func InfluxErrorJmeter() {

	ProcessInfo("Jmeter delta")
	gc := reportdata.NewInfluxClient(cfg.JmeterInflux, "", ProcessLog, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.JmeterQuery + " " + timeperiod_influx + " " + cfg.JmeterQueryGroup))

	if err != nil {
		ProcessError("InfluxErrorJmeter error")
		ProcessError(err)
		return
	}

	for _, i := range infjson.Results[0].Series {

		ProcessInfo("Test " + i.Tags.Suite + " load")

		var LTTest_yfield []reportdata.YField
		var LTTest_dinamictmp reportdata.LTTestDinamic
		var LTTest_yfieldtmp reportdata.YField

		LTTest_dinamictmp.NameTest = i.Tags.Suite

		num := 1
		for _, j := range cfg.JmeterQueryField {

			LTTest_yfieldtmp.Name = j.Name
			LTTest_yfieldtmp.Description = j.Description
			LTTest_yfieldtmp.Value = reportdata.JsonINfluxFiledParseFloat(i.Values[0][num])
			num++
			LTTest_yfield = append(LTTest_yfield, LTTest_yfieldtmp)
		}

		LTTest_dinamictmp.Field = LTTest_yfield
		LTTest_dinamic = append(LTTest_dinamic, LTTest_dinamictmp)

		ProcessDebug(LTTest_dinamictmp)
	}
}

func InfluxJmeterScenario() {

	ProcessInfo("Jmeter Scenario")
	gc := reportdata.NewInfluxClient(cfg.JmeterInflux, "", ProcessLog, debugm)
	infjson, err := gc.GetDataMean(url.QueryEscape(cfg.JmeterQueryScenario + timeperiod_influx + cfg.JmeterQueryScnrGroup))

	if err != nil {
		ProcessError("InfluxJmeterScenario error")
		ProcessError(err)
		return
	}
	var JMeterTestTh map[string]map[string]reportdata.KeyField
	JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

	LTScen_dimanict = make(map[string]map[string]reportdata.ScenarioDinamic)

	// Построение карты порогов
	ProcessInfo("Load map threshold for Scenario ")
	for _, j := range cfg.JmeterQueryScnrThreshold {
		JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name+":"+j.NameThread, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: j.Statut})
		ProcessDebug(j.Name + ":" + j.NameThread + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	for _, i := range infjson.Results[0].Series {
		if i.Tags.Statut != "" && i.Tags.Transaction != "all" {

			ProcessDebug("Load scenario " + i.Tags.Application + ":" + i.Tags.Transaction + " for statut " + i.Tags.Statut)
			// можно так но все равно потом цикл в цикле :(
			LTScenTmpt := LTScen_dimanict[i.Tags.Application][i.Tags.Transaction]
			LTScenTmpt.SetApplication(i.Tags.Application)
			LTScenTmpt.SetThread(i.Tags.Transaction)
			LTScenTmpt.SeField(reportdata.InfluxJmeterScenarioStatut(i.Values, i.Tags.Statut, cfg.JmeterQueryScnrField))
			LTScen_dimanict = reportdata.AddMapS(LTScen_dimanict, i.Tags.Application, i.Tags.Transaction, LTScenTmpt)
		}

	}

	for _, k := range LTScen_dimanict {
		for _, ii := range k {

			// Есть ли пороги для данного теста
			if val, ok := JMeterTestTh[ii.NameTest+":"+ii.NameThread]; ok {
				ProcessDebug("Load stack error for " + ii.NameTest + ":" + ii.NameThread)
				ProcessDebug(val)

				for _, jj := range ii.Field {
					// Берем имя поля с порогом
					// Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
					// Если есть определяем порог
					if valthershold, okk := JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name]; okk {

						if jj.Value > float64(valthershold.Value) && jj.Statut == JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name].Statut {
							ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
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
							ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
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
							ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
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

	ProcessInfo("Load grafana metrics")

	var p reportdata.LTGrag
	for _, i := range cfg.Grafanadash {

		ProcessInfo("Load grafana " + i.Name)
		request := i.Urlimg + timewrap
		ProcessDebug("Get image request " + request)

		gc := reportdata.NewGrafanaClient(request, i.AuthHeader, ProcessLog, debugm)
		ConType, err := gc.GetImage(i.Name)
		defer gc.Close()

		if err != nil {
			ProcessError("Error generate image")
		}

		var percentile float64
		if i.SourceType == 2 {
			//  получение данные из прометеуса
			gcs := reportdata.NewPrometheusClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
			percentile, err = gcs.GetDataSourceThreshold(url.QueryEscape(i.Query+" "+i.UrlQueryGroup) + timeperiod_prometheus)
			defer gcs.Close()
		} else if i.SourceType == 3 {
			gh := reportdata.NewGraphiteClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
			percentile, err = gh.Get99thPercentile(i.Query, timeperiodstart, timeperiodend)
			defer gh.Close()
		} else {
			//  получение данные из инфлюкса
			gcs := reportdata.NewInfluxClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
			percentile, err = gcs.GetDataSourceThreshold(url.QueryEscape(i.Query + " AND " + timeperiod_influx + i.UrlQueryGroup))
			defer gcs.Close()
		}

		if percentile > float64(i.Threshold) && err == nil {
			p := reportdata.LTError{Name: "Grafana: " + i.Name, Threshold: i.Threshold, Description: i.ThDescription + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)) + "", Type: "Grafana"}
			Problems = append(Problems, p)
		}

		p.Name = i.Name
		p.Threshold = i.Threshold
		p.ContentType = ConType
		p.Size.Height = i.Size.Height
		p.Size.Width = i.Size.Width
		p.UrlDash = i.Urldash + timewrap
		LTGrafs = append(LTGrafs, p)
	}

	GrafanaTemplateReport()
}

func GrafanaTemplateReport() {
	ProcessInfo("Load grafana template metrics")

	for _, i := range cfg.GrafanadashTemplate {
		if i.FileList != "" {
			f, err := os.Open(i.FileList)
			if err != nil {
				ProcessErrorAny("Unable to read input file "+i.FileList, err)
				ProcessError("Thread " + i.Name + " not start")
			} else {
				defer f.Close()
				ProcessDebug("Start load " + i.FileList)

				// read csv values using csv.Reader
				csvReader := csv.NewReader(f)
				// Разделитель CSV
				csvReader.Comma = ';'
				csv, err := csvReader.ReadAll()

				if err != nil {
					ProcessError(err)
					continue
				}

				var result []reportdata.DinamicRecord

				headers := strings.Split(i.FilePettern, ";")
				colunmlen := len(headers)

				for j, line := range csv {
					// Проверяем соответствие количества столбцов
					if len(line) != len(headers) {
						ProcessError(fmt.Errorf("Row %d: does not match the number of columns ", j+1))
						continue
					}
					record := make(reportdata.DinamicRecord, colunmlen)

					tmp := i.Urldash
					tmp_query := i.Query
					for jj, head := range headers {
						record[head] = line[jj]
						tmp = strings.Replace(tmp, "{"+head+"}", line[jj], 1)
						tmp_query = strings.Replace(tmp_query, "{"+head+"}", line[jj], 1)
					}
					result = append(result, record)

					var percentile float64
					if i.SourceType == 2 {
						//  получение данные из прометеуса
						gcs := reportdata.NewPrometheusClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
						percentile, err = gcs.GetDataSourceThreshold(url.QueryEscape(i.Query+" "+i.UrlQueryGroup) + timeperiod_prometheus)
						defer gcs.Close()
					} else if i.SourceType == 3 {
						gh := reportdata.NewGraphiteClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
						percentile, err = gh.Get99thPercentile(tmp_query, timeperiodstart, timeperiodend)
						defer gh.Close()
					} else {
						//  получение данные из инфлюкса
						gcs := reportdata.NewInfluxClient(i.UrlQuery, i.AuthHeader, ProcessLog, debugm)
						percentile, err = gcs.GetDataSourceThreshold(url.QueryEscape(i.Query + " AND " + timeperiod_influx + i.UrlQueryGroup))
						defer gcs.Close()
					}
					if percentile > float64(i.Threshold) && err == nil {
						ProcessDebug(i.Name + " " + strings.Join(line, ", ") + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)))
						p := reportdata.LTError{Name: "Grafana: " + i.Name + " " + strings.Join(line, ", "), Threshold: i.Threshold, Description: i.ThDescription + " " + strings.Join(line, ", ") + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)) + "", Type: "Grafana"}
						Problems = append(Problems, p)
					} else if err != nil {
						ProcessError(err)
					}
				}
			}
		} else {
			ProcessInfo("Pool not defined for " + i.Name + " template dash")
		}
	}
}

func ClickHouseReport() {
	ProcessInfo("Start load ClickHouse")

	ch := reportdata.NewCHClient("http://"+cfg.ClickHouse.Server+"/?", cfg.ClickHouse.User, cfg.ClickHouse.Pass, ProcessLog, debugm)
	for _, i := range cfg.ClickHouse.Query {
		clkhouse, err := ch.GetSql(i.DBname, i.Sql, i.Name, timeperiod_clickhouse)
		if err == nil {
			LTClickHouse = append(LTClickHouse, clkhouse)
		}
	}
	defer ch.Close()
}

func ReportIM() {
	/*	client := &http.Client{}

		resp, err := client.Get("https:// "+cfg.FSMConnect+"/incidents?")

		ProcessInfo(resp)

		if err != nil {
			ProcessError(err)

		}*/
	ProcessInfo("Start report FSM")

}

//  Загрузка в джиру
func ReportDownload(reportfilename string) {
	// пробрасываем дебаг
	confluence.DebugFlag = debugm
	// Указваем что писать в лог
	confluence.LogFlag = true

	// Инициализация работы с конфленсом
	// перенеммные потом вынески в конфиг
	confl, err := confluence.NewAPI(cfg.ReportConfluenceURL, cfg.ReportConfluenceLogin, cfg.ReportConfluencePass, cfg.ReportConfluenceToken, cfg.ReportConfluenceProxy)
	if err != nil {
		ProcessError("Error connection to confluence")
		ProcessError(err)
		return
	}
	// Получение описание базовой страницы
	ProcessDebug("GetContent")
	JsonCont, err := confl.GetContent(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Expand: []string{"children.page"}}, ProcessLog)
	if err != nil {
		ProcessError(err)
		ProcessError(JsonCont)
		return
	}
	ProcessDebug("GetContentChildPage")
	JsonConC, err := confl.GetContentChildPage(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Limit: 250, Expand: []string{"children.page"}}, ProcessLog)
	if err != nil {
		ProcessError(err)
		ProcessError(JsonConC)
		return
	}

	currentTime := time.Now()
	reportname := "Report" + currentTime.Format("20060102")

	var IdChild string

	// поиск по детям GetContentChildPage
	for _, i := range JsonConC.Results {
		if reportname == i.Title {
			IdChild = i.ID
			ProcessDebug(i.Title + ", id=" + IdChild)
		} else {
			IdChild = ""
		}
	}

	if IdChild == "" {
		ProcessDebug("Create child page " + reportname)
		// формирование тела для создания
		data := confluence.ConflCreateType{
			Type:  "page",
			Title: reportname,
			Ancestors: []confluence.Ancestor{
				confluence.Ancestor{
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

		JsonContC, err := confl.CreateContent(&data, ProcessLog)
		if err != nil {

			ProcessError(err)
			return
		}
		IdChild = JsonContC.ID

		file, err := os.OpenFile(reportfilename, os.O_RDONLY, 0666)
		if err != nil {
			ProcessError(err)
		}

		ProcessDebug("Upload current attachments " + reportfilename)
		arsp, err := confl.UploadAttachment(IdChild, reportfilename, file, ProcessLog)
		if err != nil {
			ProcessError(err)
			ProcessDebug(arsp)
		}

		defer file.Close()

	} else {
		ProcessDebug("Load current attachments")
		arsp, err := confl.GetAttachments(IdChild, ProcessLog)
		if err != nil {
			ProcessError(err)
			ProcessDebug(arsp)
		}

		chck := false
		attachid := "0"
		for _, j := range arsp.Results {
			if j.Title == reportfilename {
				attachid = j.ID
				chck = true
			}
		}

		file, err := os.OpenFile(reportfilename, os.O_RDONLY, 0666)
		if err != nil {
			ProcessInfo(err)
		}

		if chck {
			ProcessDebug("Update current attachments " + reportfilename)
			arsp, err := confl.UpdateAttachment(IdChild, reportfilename, attachid, file, ProcessLog)
			if err != nil {
				ProcessInfo(err)
				ProcessDebug(arsp)
			}
		} else {
			ProcessDebug("Upload current attachments " + reportfilename)
			arsp, err := confl.UploadAttachment(IdChild, reportfilename, file, ProcessLog)
			if err != nil {
				ProcessInfo(err)
				ProcessDebug(arsp)
			}
		}

		defer file.Close()

	}

}

func RemoveTemp() {

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
