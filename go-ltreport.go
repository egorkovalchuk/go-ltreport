package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egorkovalchuk/go-ltreport/confluence"
	"github.com/egorkovalchuk/go-ltreport/reportdata"

	"github.com/jung-kurt/gofpdf"
)

//Power by  Egor Kovalchuk
//0.2.0.0 Add any time period
//0.2.0.1 Add token for confluence
//0.3.0.0 Add ClickHouse
//0.3.0.1 Add PDF report CH

const (
	// логи
	logFileName  = "ltreport.log"
	confFileName = "config.json"
	versionutil  = "0.3.0.1"
	a4height     = 297
	a4width      = 210
)

var (
	//Configuration
	cfg reportdata.Config
	//FSM connect
	LoginFSM string
	PassFSM  string
	//Proxy
	//Confluence
	ConflProxy string
	ConfToken  string

	//ClickHouse
	CHUser string
	CHPass string

	//режим работы сервиса(дебаг мод)
	debugm bool
	//Запись в лог
	filer *os.File
	//запрос помощи
	help bool
	//по часовой отчет за прошедший час
	hour bool
	// Delete temp files
	rmtmpfile bool
	//PDF
	pdf *gofpdf.Fpdf
	//ошибки
	err error
	//запрос версии
	version bool
	//Переменная для тестов
	LTTests        []reportdata.LTTest
	LTTest_dinamic []reportdata.LTTestDinamic
	//Переменная для анализа
	Problems []reportdata.LTError
	//Переменная для сценариев
	LTScenario []reportdata.Scenario
	//Устарело?
	//LTScen_dimanic  map[string]reportdata.ScenarioDinamic
	LTScen_dimanict map[string]map[string]reportdata.ScenarioDinamic
	//Массив графиков и порогов
	LTGrafs []reportdata.LTGrag
	//время формирования отчетов
	timeperiod            string
	timeperiod_prometheus string
	timeperiod_clickhouse string
	//Пользовательский период формировани
	EndDateStr   string
	StartDateStr string
	EndDate      time.Time
	StartDate    time.Time

	//Массив для ClickHouse
	LTClickHouse []reportdata.ClickHouseJson
)

//чтение конфига
func readconf(cfg *reportdata.Config, confname string) {
	file, err := os.Open(confname)
	if err != nil {
		ProcessError(err)
		fmt.Println(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		ProcessError(err)
		fmt.Println(err)
	}

	file.Close()
}

func redefinitionconf() {

	//Замена переменных для FSM
	//переместить?
	if LoginFSM != "" {
		cfg.ReportIM.LoginFSM = LoginFSM
	}
	if PassFSM != "" {
		cfg.ReportIM.PassFSM = PassFSM
	}

	//Замена переменных Proxy
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

//Запись ошибки с прекращением выполнения
func ProcessError(err error) {
	log.Println(err)
	os.Exit(2)
}

//Запись в лог при включенном дебаге
func ProcessDebug(logtext interface{}) {
	if debugm {
		log.Println(logtext)
	}
}

//Применение шаблона к датам по произвольному периоду
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

func main() {

	//start program
	filer, err = os.OpenFile(logFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(filer)
	log.Println("- - - - - - - - - - - - - - -")
	log.Println("Start report")

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
	//Замена на приоритетный конфиг из командной строки
	redefinitionconf()

	//Получение помощи
	if help {
		reportdata.Helpstart()
		return
	}

	//получение версии
	if version {
		fmt.Println("Version utils " + versionutil)
		return
	}

	//Проверка по датам
	if !DateProcess() {
		return
	}

	ProcessDebug("Start with debug mode")
	StartReport()
	RemoveTemp()

}

func StartReport() {

	currentTime := time.Now()

	//Формирование имени выхожного файла
	reportfilename := cfg.ReportFilename + currentTime.Format(cfg.ReportMask)

	if hour {
		log.Println("Start group by hour")
		timeperiod = ` time >= ` + strconv.FormatInt(reportdata.BeginningOfHour().Unix(), 10) + `000ms AND time <= ` + strconv.FormatInt(reportdata.EndOfHour().Unix(), 10) + `000ms `
		timeperiod_prometheus = `&start=` + strconv.FormatInt(reportdata.BeginningOfHour().Unix(), 10) + `&end=` + strconv.FormatInt(reportdata.EndOfHour().Unix(), 10)
		timeperiod_clickhouse = " timestamp>=toDateTime('" + reportdata.BeginningOfHour().Format("2006-01-02 15:04:05") + "') and timestamp <=toDateTime('" + reportdata.EndOfHour().Format("2006-01-02 15:04:05") + "') "
		reportfilename = reportfilename + "_" + strconv.Itoa(reportdata.BeginningOfHour().Local().Hour())
	} else if !StartDate.IsZero() && !EndDate.IsZero() {
		log.Println("Start group arbitrary period")
		timeperiod = ` time >= ` + strconv.FormatInt(StartDate.Unix(), 10) + `000ms AND time <= ` + strconv.FormatInt(EndDate.Unix(), 10) + `000ms `
		timeperiod_prometheus = `&start=` + strconv.FormatInt(StartDate.Unix(), 10) + `&end=` + strconv.FormatInt(EndDate.Unix(), 10)
		timeperiod_clickhouse = " timestamp>=toDateTime('" + StartDate.Format("2006-01-02 15:04:05") + "') and timestamp <=toDateTime('" + EndDate.Format("2006-01-02 15:04:05") + "') "
		reportfilename = reportfilename + "_" + "00" + "_" + strconv.Itoa(StartDate.Local().Hour())
	} else {
		log.Println("Start group by day")
		timeperiod = ` time >= ` + strconv.FormatInt(reportdata.BeginningOfDay().Unix(), 10) + `000ms AND time <= ` + strconv.FormatInt(reportdata.EndOfDay().Unix(), 10) + `000ms `
		timeperiod_prometheus = `&start=` + strconv.FormatInt(reportdata.BeginningOfDay().Unix(), 10) + `&end=` + strconv.FormatInt(reportdata.EndOfDay().Unix(), 10)
		timeperiod_clickhouse = " timestamp>=toDateTime('" + reportdata.BeginningOfDay().Format("2006-01-02 15:04:05") + "') and timestamp <=toDateTime('" + reportdata.EndOfDay().Format("2006-01-02 15:04:05") + "') "
	}

	fmt.Println("Start report")
	log.Println("Start report")
	ProcessDebug("File name " + reportfilename)

	//Инициализация pdf
	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report load testing for "+currentTime.Format("02.01.2006"), "1", 0, "CM", false, 0, "")

	//Получение инцидентов
	if cfg.ReportOn.ReportIM {
		ReportIM()
	}
	//заведение работ
	//Загрузка данных из графаны
	if cfg.ReportOn.ReportJmeter {
		ReportInflux()
	}
	//Загрузка данных из графаны по сценариям
	if cfg.ReportOn.ReportScenario {
		InfluxJmeterScenario()
	}

	if cfg.ReportOn.ReportDash {
		GrafanaReport()
	}

	//Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		ClickHouseReport()
	}

	//Формирование отчета
	//Обязательно, поэтому не исключаем
	ReportProblemPDF()

	//Включение графиков
	if cfg.ReportOn.ReportDash {
		GrafanaReportPDF()
	}

	//Включение ClickHouse
	if cfg.ReportOn.ReportClickHouse {
		ClickHouseReportPDF()
	}

	//Прогрузка общенй информации
	if cfg.ReportOn.ReportJmeter {
		ReportInfluxPDF()
		ReportProblemScenPDF()
		ReportInfluxScrnPDF()
	}

	//сжатие
	pdf.SetCompression(true)
	//запись отчета
	pdf.OutputFileAndClose(reportfilename + ".pdf")

	if cfg.ReportConfluenceOn {
		ReportDownload(reportfilename + ".pdf")
	}

}

func ReportProblemPDF() {
	ProcessDebug("Start generate pdf - Problem ")
	for _, i := range Problems {
		if i.Type == "Grafana" || i.Type == "" {
			pdf.SetFont("Times", "", 10)
			pdf.SetY(pdf.GetY() + 7)
			pdf.CellFormat(65, 7, i.Description, "", 0, "LM", false, 0, "")
		}
	}
}

func ReportProblemScenPDF() {
	ProcessDebug("Start generate pdf - Problem scenario")
	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report problem Jmeter", "0", 0, "CM", false, 0, "")
	pdf.SetY(pdf.GetY() + 6)

	for _, i := range Problems {
		if i.Type == "Jmeter" {
			pdf.SetFont("Times", "", 6)
			pdf.SetY(pdf.GetY() + 3)
			pdf.CellFormat(65, 3, i.Description, "", 0, "LM", false, 0, "")
		}
	}
}
func ReportInfluxPDF() {
	ProcessDebug("Start generate pdf - Influx JMeter")
	pdf.AddPage()

	pdf.SetY(pdf.GetY() + 10)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report test", "0", 0, "CM", false, 0, "")
	pdf.SetY(pdf.GetY() + 7)

	for _, i := range LTTest_dinamic {
		pdf.SetFont("Times", "B", 10)
		pdf.CellFormat(195, 7, i.NameTest, "0", 0, "CM", false, 0, "")

		pdf.SetFont("Times", "", 8)
		pdf.SetY(pdf.GetY() + 7)

		num := 1
		for _, ii := range i.Field {

			pdf.CellFormat(65, 7, fmt.Sprintf(ii.Description, int(ii.Value)), "1", 0, "LM", false, 0, "")

			if num%3 == 0 {
				pdf.SetY(pdf.GetY() + 7)
			}
			num = num + 1
		}

	}

}

func ReportInfluxScrnPDF() {
	ProcessDebug("Start generate pdf - Influx Scenario Jmeter")
	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 10)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report scenario", "0", 0, "CM", false, 0, "")

	var keys []string
	for k := range LTScen_dimanict {
		keys = append(keys, k)
	}

	for _, k := range keys {

		pdf.SetY(pdf.GetY() + 6)
		pdf.SetFont("Times", "B", 8)
		pdf.CellFormat(195, 7, k, "0", 0, "CM", false, 0, "")

		for _, i := range LTScen_dimanict[k] {
			//pdf.SetY(pdf.GetY() + 6)
			pdf.SetFont("Times", "B", 6)
			pdf.CellFormat(195, 7, i.NameTest+":"+i.NameThread, "0", 0, "CM", false, 0, "")
			pdf.SetY(pdf.GetY() + 6)

			scenariopdf := make(map[string][]reportdata.YField)
			//Для красивого вывода в отчет
			for _, ii := range i.Field {
				scenariopdf[ii.Statut] = append(scenariopdf[ii.Statut], ii)
			}

			if _, ok := scenariopdf["all"]; ok {
				pdf.SetFont("Times", "", 4)
				pdf.CellFormat(20, 4, "All ", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["all"] {
					pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						pdf.SetY(pdf.GetY() + 4)
					}
					num++
				}
			}

			if _, ok := scenariopdf["ok"]; ok {

				pdf.CellFormat(20, 4, "Status OK", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["ok"] {
					pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						pdf.SetY(pdf.GetY() + 4)
					}
					num++
				}
			}

			if _, ok := scenariopdf["ko"]; ok {

				pdf.CellFormat(20, 4, "Error ", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["ko"] {
					pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						pdf.SetY(pdf.GetY() + 4)
					}
					num++
				}
			}

		}
	}

}

func GrafanaReportPDF() {
	ProcessDebug("Start generate pdf - Grafana")
	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)

	var saveX, saveY, tmpheight float64
	saveX = 10
	tmpheight = 0
	saveY = 6

	for _, i := range LTGrafs {

		//сохранение переменных для второго и послед рисунков

		ln := pdf.PointConvert(8)

		if saveX+float64(i.Size.Width) > a4width {
			pdf.SetXY(10, saveY+tmpheight+ln+2)
			saveX = 10
			saveY = saveY + tmpheight + ln + 2
		}

		if pdf.GetY()+ln+float64(i.Size.Height)+7 > a4height || saveY > a4height {
			pdf.AddPage()
			saveY = 6
		}

		pdf.SetFont("Times", "B", 8)
		pdf.SetXY(saveX, saveY)
		pdf.MultiCell(float64(i.Size.Width), ln, i.Name, "", "C", false)

		tp := pdf.ImageTypeFromMime(i.ContentType)

		file, err := os.Open(i.Name + ".png")
		if err != nil {
			log.Println(err)
		}

		_ = pdf.RegisterImageOptionsReader(i.Name, gofpdf.ImageOptions{ImageType: tp, ReadDpi: false, AllowNegativePosition: true}, file)
		if pdf.Ok() {

			pdf.Image(i.Name, saveX, saveY+4,
				float64(i.Size.Width), float64(i.Size.Height), false, "png", 0, i.UrlDash)
			saveX = saveX + float64(i.Size.Width) + 10

		}

		tmpheight = float64(i.Size.Height)

		defer file.Close()
		file.Close()

		if rmtmpfile {
			e := os.Remove(file.Name())
			if e != nil {
				log.Println(err)
			}
		}

	}
}

func ClickHouseReportPDF() {
	ProcessDebug("Start generate pdf - ClickHouse")
	var saveX, saveY, tmpheight float64
	saveX = 6
	tmpheight = 0
	saveY = 10

	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report ClickHouse", "0", 0, "CM", false, 0, "")
	saveX, saveY = pdf.GetXY()

	//добавить вычисление длины для таблиц

	for _, i := range LTClickHouse {

		ln := pdf.PointConvert(6)

		if pdf.GetY()+float64(25)+ln > a4height || saveY+25+ln > a4height {
			pdf.AddPage()
			saveY = 10
			saveX = 6
		}

		pdf.SetXY(saveX, saveY)
		pdf.SetY(pdf.GetY() + 8)
		pdf.SetFont("Times", "B", 10)
		pdf.CellFormat(195, 7, i.Name, "0", 0, "CM", false, 0, "")

		pdf.SetY(pdf.GetY() + 8)
		saveX, saveY = pdf.GetXY()

		// При создании новой таблицы проверяем, что она влазит на страницу
		// Добавить перенос строк
		// разобраться с отступом
		saveX = 6

		for _, j := range i.Meta {
			pdf.SetFont("Times", "B", 5)
			pdf.Rect(saveX, saveY, float64(j.Len+4), 5+tmpheight, "")
			pdf.MultiCell(float64(j.Len+4), 5+tmpheight, j.Name, "", "CM", false)
			saveX += float64(j.Len + 4)
			pdf.SetXY(saveX, saveY)
		}

		for _, jj := range i.Data {
			saveX = 6
			saveY += 5

			if pdf.GetY()+float64(25)+ln > a4height || saveY+25+ln > a4height {
				pdf.AddPage()
				saveX = 6
				saveY = 10
			}

			pdf.SetXY(saveX, saveY)
			pdf.SetFont("Times", "", 5)

			for _, jjj := range i.Meta {
				pdf.Rect(saveX, saveY, float64(jjj.Len+4), 5+tmpheight, "")
				pdf.MultiCell(float64(jjj.Len+4), 5+tmpheight, jj[jjj.Name].(string), "", "", false)
				saveX += float64(jjj.Len + 4)
				pdf.SetXY(saveX, saveY)
			}

		}
		saveX, saveY = pdf.GetXY()
	}
}

func ReportInflux() {

	//Получение данных из инфлюкса jmeter
	InfluxErrorJmeter()

	var JMeterTestTh map[string]map[string]reportdata.KeyField
	JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

	//Построение карты порогов
	ProcessDebug("Load map threshold for tests ")
	for _, j := range cfg.JmeterQueryThreshold {
		JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: ""})
		ProcessDebug(j.Name + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
	}

	ProcessDebug("Report generation ")

	//Формирование стека ошибок
	for _, ii := range LTTest_dinamic {

		ProcessDebug("Load stack error for " + ii.NameTest)

		//Есть ли пороги для данного теста
		if val, ok := JMeterTestTh[ii.NameTest]; ok {
			ProcessDebug(val)

			for _, jj := range ii.Field {
				//Берем имя поля с порогом
				//Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				//Если есть определяем порог
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
				//Берем имя поля с порогом
				//Проверяем, есть ли порог для теста ii.NameTest c полем jj.Name
				//Если есть определяем порог
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

	request := cfg.JmeterInflux + url.QueryEscape(cfg.JmeterQuery+" "+timeperiod+" "+cfg.JmeterQueryGroup)
	ProcessDebug(request)

	resp, err := http.Get(request)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {

		log.Println("HTTP Status is in the 2xx range " + request)

		infjson, err := reportdata.JsonINfluxParse(resp)

		for _, i := range infjson.Results[0].Series {

			if err == nil {

				ProcessDebug("Test " + i.Tags.Suite + " load")

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

				if debugm {
					log.Println(LTTest_dinamictmp)
				}

			}

		}
	} else {
		resp.Body.Close()
		log.Println("HTTP Status error " + strconv.Itoa(resp.StatusCode) + " " + request)
		ProcessDebug(fmt.Errorf("Response Status: %s", resp.Status))
	}
}

func InfluxJmeterScenario() {

	request := cfg.JmeterInflux + url.QueryEscape(cfg.JmeterQueryScenario+timeperiod+cfg.JmeterQueryScnrGroup)

	ProcessDebug("JmeterScenario")
	ProcessDebug(request)

	resp, err := http.Get(request)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {

		log.Println("HTTP Status is in the 2xx range " + request)

		infjson, _ := reportdata.JsonINfluxParse(resp)

		resp.Body.Close()

		var JMeterTestTh map[string]map[string]reportdata.KeyField
		JMeterTestTh = make(map[string]map[string]reportdata.KeyField)

		// Устарело?
		//LTScen_dimanic = make(map[string]reportdata.ScenarioDinamic)
		LTScen_dimanict = make(map[string]map[string]reportdata.ScenarioDinamic)

		//Построение карты порогов
		ProcessDebug("Load map threshold for Scenario ")
		for _, j := range cfg.JmeterQueryScnrThreshold {
			JMeterTestTh = reportdata.AddMap(JMeterTestTh, j.Name+":"+j.NameThread, j.ErrorField, reportdata.KeyField{Value: j.Threshold, Description: j.Description, Statut: j.Statut})
			ProcessDebug(j.Name + ":" + j.NameThread + " threshold " + strconv.Itoa(JMeterTestTh[j.Name][j.ErrorField].Value) + " for field " + j.ErrorField)
		}

		for _, i := range infjson.Results[0].Series {
			if i.Tags.Statut != "" && i.Tags.Transaction != "all" {

				ProcessDebug("Load scenario " + i.Tags.Application + ":" + i.Tags.Transaction + " for statut " + i.Tags.Statut)
				/* Устарело?
				//Сохраняем текущие настройки
				LTScenTmp := LTScen_dimanic[i.Tags.Application+":"+i.Tags.Transaction]
				LTScenTmp.SetApplication(i.Tags.Application)
				LTScenTmp.SetThread(i.Tags.Transaction)
				LTScenTmp.SeField(InfluxJmeterScenarioStatut(i.Values, i.Tags.Statut))
				//Сохраняем в карту
				LTScen_dimanic[i.Tags.Application+":"+i.Tags.Transaction] = LTScenTmp
				*/

				//можно так но все равно потом цикл в цикле :(
				LTScenTmpt := LTScen_dimanict[i.Tags.Application][i.Tags.Transaction]
				LTScenTmpt.SetApplication(i.Tags.Application)
				LTScenTmpt.SetThread(i.Tags.Transaction)
				LTScenTmpt.SeField(reportdata.InfluxJmeterScenarioStatut(i.Values, i.Tags.Statut, cfg.JmeterQueryScnrField))
				LTScen_dimanict = reportdata.AddMapS(LTScen_dimanict, i.Tags.Application, i.Tags.Transaction, LTScenTmpt)
			}

		}

		for _, k := range LTScen_dimanict {
			for _, ii := range k {

				//Есть ли пороги для данного теста
				if val, ok := JMeterTestTh[ii.NameTest+":"+ii.NameThread]; ok {
					ProcessDebug("Load stack error for " + ii.NameTest + ":" + ii.NameThread)
					ProcessDebug(val)

					for _, jj := range ii.Field {
						//Берем имя поля с порогом
						//Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
						//Если есть определяем порог
						if valthershold, okk := JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name]; okk {

							if jj.Value > float64(valthershold.Value) && jj.Statut == JMeterTestTh[ii.NameTest+":"+ii.NameThread][jj.Name].Statut {
								ProcessDebug(ii.NameTest + ":" + ii.NameThread + " threshold: " + strconv.Itoa(valthershold.Value) + " current: " + strconv.Itoa(int(jj.Value)))
								p := reportdata.LTError{Name: ii.NameTest + ":" + ii.NameThread, Threshold: valthershold.Value, Description: fmt.Sprintf(valthershold.Description, valthershold.Value, ii.NameTest+":"+ii.NameThread, int(jj.Value)), Type: "Jmeter"}
								Problems = append(Problems, p)

							}
						}

					}
					//смотрим дефолты
				} else if _, ok := JMeterTestTh["*:*"]; ok {

					for _, jj := range ii.Field {
						//Берем имя поля с порогом
						//Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
						//Если есть определяем порог
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
						//Берем имя поля с порогом
						//Проверяем, есть ли порог для теста ii.Tags c полем jj.Name
						//Если есть определяем порог
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
	} else {
		resp.Body.Close()
		log.Println("HTTP Status error " + strconv.Itoa(resp.StatusCode) + " " + request)
		ProcessDebug(fmt.Errorf("Response Status: %s", resp.Status))
	}
}

func GrafanaReport() {

	ProcessDebug("Load grafana metrics")

	var p reportdata.LTGrag
	for _, i := range cfg.Grafanadash {

		var beginday int64
		var endday int64

		beginday = reportdata.BeginningOfDay().Unix()
		endday = reportdata.EndOfDay().Unix()

		ProcessDebug("Load grafana " + i.Name)
		var timewrap string
		if hour {
			timewrap = "&from=" + strconv.FormatInt(reportdata.BeginningOfHour().Unix(), 10) + "000&to=" + strconv.FormatInt(reportdata.EndOfHour().Unix(), 10) + "000"
		} else if !StartDate.IsZero() && !EndDate.IsZero() {
			timewrap = "&from=" + strconv.FormatInt(StartDate.Unix(), 10) + "000&to=" + strconv.FormatInt(EndDate.Unix(), 10) + "000"
		} else {
			timewrap = "&from=" + strconv.FormatInt(beginday, 10) + "000&to=" + strconv.FormatInt(endday, 10) + "000"
		}

		request := i.Urlimg + timewrap
		ProcessDebug("Request " + request)

		resp, err := http.NewRequest("GET", request, nil)
		resp.Header.Add("Authorization", i.AuthHeader)
		resp.Header.Add("Content-Type", "image/jpeg")

		if err != nil {
			log.Println(err)
		}

		cli := &http.Client{}
		rsp, err := cli.Do(resp)

		if err != nil {
			log.Println(err)
		}

		// проверяем получение картинки, статус 200
		if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {

			log.Println("Request success")

			var n io.Reader
			//io.Copy(ioutil.Discard, rsp.Body)
			nn, err := ioutil.ReadAll(rsp.Body)
			n = bytes.NewReader(nn)

			if err != nil {
				log.Println(err)
			}

			//open a file for writing
			file, err := os.Create(i.Name + ".png")
			if err != nil {
				log.Fatal(err)
			}

			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, n)
			if err != nil {
				log.Fatal(err)
			}

			defer rsp.Body.Close()
			defer file.Close()

			request_inf := ""
			if i.SourceType == 2 {
				// получение данные из прометеуса
				request_inf = i.UrlQuery + url.QueryEscape(i.Query+" "+i.UrlQueryGroup) + timeperiod_prometheus
			} else {
				// получение данные из инфлюкса
				request_inf = i.UrlQuery + url.QueryEscape(i.Query+" AND "+timeperiod+i.UrlQueryGroup)
			}
			ProcessDebug("Request image " + request_inf)

			resp_inf, err := http.NewRequest("GET", request_inf, nil)
			resp_inf.Header.Add("Authorization", i.AuthHeader)

			if err != nil {
				log.Println(err)
			}

			cli_inf := &http.Client{}
			rsp_inf, err := cli_inf.Do(resp_inf)

			if err != nil {
				log.Println(err)
			}

			var percentile float64
			if rsp_inf.StatusCode >= 200 && rsp_inf.StatusCode <= 299 {

				log.Println("Request Grafana threshold success")

				if i.SourceType == 2 {
					var prom reportdata.PrometheusResponse
					err = prom.JsonPrometheusParse(rsp_inf)
					if err == nil {
						percentile = prom.JsonPrometheusFiledParseFloat(prom.Data.Result[0].Value[1])
					} else {
						log.Println(err)
					}
				} else {
					infjson_child, err := reportdata.JsonINfluxParse(rsp_inf)
					if err == nil {
						percentile = reportdata.JsonINfluxFiledParseFloat(infjson_child.Results[0].Series[0].Values[0][1])
					}
				}

				if percentile > float64(i.Threshold) {
					p := reportdata.LTError{Name: "Grafana: " + i.Name, Threshold: i.Threshold, Description: i.ThDescription + ": Threshold " + strconv.Itoa(i.Threshold) + " - current " + strconv.Itoa(int(percentile)) + "", Type: "Grafana"}
					Problems = append(Problems, p)
				}

				p.Name = i.Name
				p.Threshold = i.Threshold
				p.ContentType = rsp.Header["Content-Type"][0]
				p.Size.Height = i.Size.Height
				p.Size.Width = i.Size.Width
				p.UrlDash = i.Urldash + timewrap
				LTGrafs = append(LTGrafs, p)

			} else {
				log.Println("Request Grafana threshold  error " + strconv.Itoa(rsp_inf.StatusCode) + " " + request_inf)
			}
		} else {
			rsp.Body.Close()
			ProcessDebug(fmt.Errorf("Response Status: %s", rsp.Status))
			log.Println("Request error " + strconv.Itoa(rsp.StatusCode) + " " + request)
		}
	}

}

func ClickHouseReport() {
	ProcessDebug("Start load ClickHouse")

	//rezlen := len(cfg.ClickHouse.Query)

	for _, i := range cfg.ClickHouse.Query {
		request := "http://" + cfg.ClickHouse.Server + "/?"

		resp, err := http.NewRequest("GET", request, nil)
		if err != nil {
			log.Println(err)
		}
		ProcessDebug(strings.Replace(i.Sql, "{timestamp}", timeperiod_clickhouse, 1) + " FORMAT JSONStrings")

		resp.SetBasicAuth(cfg.ClickHouse.User, cfg.ClickHouse.Pass)
		resp.Header.Add("Content-Type", "application/json")
		resp.Header.Add("X-ClickHouse-Progress", "1")
		resp.Header.Add("X-ClickHouse-Database", i.DBname)
		resp.Header.Add("User-Agent", "go-LT-Report")
		resp.Body = ioutil.NopCloser(strings.NewReader(strings.Replace(i.Sql, "{timestamp}", timeperiod_clickhouse, 1) + " FORMAT JSONStrings"))

		cli := &http.Client{}
		rsp, err := cli.Do(resp)

		if err != nil {
			log.Println(err)
		}

		if rsp.StatusCode >= 200 && rsp.StatusCode <= 299 {
			log.Println("Query ClickHouse succes ")

			var clkhouse reportdata.ClickHouseJson
			err = clkhouse.JsonClickHouseParse(rsp, i.Name)
			LTClickHouse = append(LTClickHouse, clkhouse)

		} else {
			log.Println("Query ClickHouse error " + strconv.Itoa(rsp.StatusCode) + " " + request)
		}

		defer cli.CloseIdleConnections()

	}

}

func ReportIM() {
	/*	client := &http.Client{}

		resp, err := client.Get("https://"+cfg.FSMConnect+"/incidents?")

		log.Println(resp)

		if err != nil {
			ProcessError(err)

		}*/
	ProcessDebug("Start report FSM")

}

// Загрузка в джиру
func ReportDownload(reportfilename string) {
	//пробрасываем дебаг
	confluence.DebugFlag = debugm
	//Указваем что писать в лог
	confluence.LogFlag = true

	//Инициализация работы с конфленсом
	//перенеммные потом вынески в конфиг
	confl, err := confluence.NewAPI(cfg.ReportConfluenceURL, cfg.ReportConfluenceLogin, cfg.ReportConfluencePass, cfg.ReportConfluenceToken, cfg.ReportConfluenceProxy)
	if err != nil {
		log.Println("Error connection to confluence")
		log.Println(err)
		return
	}
	//Получение описание базовой страницы
	ProcessDebug("GetContent")
	JsonCont, err := confl.GetContent(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Expand: []string{"children.page"}})
	if err != nil {
		log.Println(err)
		log.Println(JsonCont)
		return
	}
	ProcessDebug("GetContentChildPage")
	JsonConC, err := confl.GetContentChildPage(cfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: cfg.ReportConfluenceSpace, Limit: 250, Expand: []string{"children.page"}})
	if err != nil {
		log.Println(err)
		log.Println(JsonConC)
		return
	}

	currentTime := time.Now()
	reportname := "Report" + currentTime.Format("20060102")

	var IdChild string

	//поиск по детям GetContentChildPage
	for _, i := range JsonConC.Results {
		if reportname == i.Title {
			IdChild = i.ID
			ProcessDebug(i.Title + ", id=" + IdChild)
		} else {
			IdChild = ""
		}
	}

	//Выключил, так как прямой поиск по детям
	/*for _, i := range JsonCont.Children.Page.Results {
		if reportname == i.Title {
			IdChild = i.ID
			ProcessDebug(i.Title + ", id=" + IdChild)
		} else {
			IdChild = ""
		}
	}*/

	if IdChild == "" {
		ProcessDebug("Create child page " + reportname)
		//формирование тела для создания
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

		JsonContC, err := confl.CreateContent(&data)
		if err != nil {

			log.Println(err)
			return
		}
		IdChild = JsonContC.ID

		file, err := os.OpenFile(reportfilename, os.O_RDONLY, 0666)
		if err != nil {
			log.Println(err)
		}

		ProcessDebug("Upload current attachments " + reportfilename)
		arsp, err := confl.UploadAttachment(IdChild, reportfilename, file)
		if err != nil {
			log.Println(err)
			ProcessDebug(arsp)
		}

		defer file.Close()

	} else {
		ProcessDebug("Load current attachments")
		arsp, err := confl.GetAttachments(IdChild)
		if err != nil {
			log.Println(err)
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
			log.Println(err)
		}

		if chck {
			ProcessDebug("Update current attachments " + reportfilename)
			arsp, err := confl.UpdateAttachment(IdChild, reportfilename, attachid, file)
			if err != nil {
				log.Println(err)
				ProcessDebug(arsp)
			}
		} else {
			ProcessDebug("Upload current attachments " + reportfilename)
			arsp, err := confl.UploadAttachment(IdChild, reportfilename, file)
			if err != nil {
				log.Println(err)
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
