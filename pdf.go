package main

import (
	"fmt"
	"os"

	"github.com/egorkovalchuk/go-ltreport/reportdata"

	"github.com/jung-kurt/gofpdf"
)

var (
	// PDF
	pdf            *gofpdf.Fpdf
	reportfilename string
)

func ReportInit() {

	ProcessInfo("File name " + reportfilename)
	// Инициализация pdf
	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report load testing for "+timeperiodstart.Format("02.01.2006 15:04")+"-"+timeperiodend.Format("02.01.2006 15:04"), "1", 0, "CM", false, 0, "")
}

func ReportEnd() {
	// сжатие
	pdf.SetCompression(true)
	// запись отчета
	pdf.OutputFileAndClose(reportfilename + ".pdf")
	if err != nil {
		ProcessError(err)
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
			ProcessError(err)
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
				ProcessInfo(err)
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
