package main

import (
	"fmt"
	"os"

	"github.com/egorkovalchuk/go-ltreport/internal/fonts"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"

	"github.com/jung-kurt/gofpdf"
)

var (
	// PDF
	pdf            *gofpdf.Fpdf
	reportfilename string
)

func ReportPDFInit() {

	fontManager := fonts.NewFontManager()
	defer fontManager.Cleanup()

	logs.ProcessInfo("File name " + reportfilename)
	// Инициализация pdf
	pdf = gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Настраиваем шрифты
	if err := fontManager.SetupFonts(pdf); err != nil {
		logs.ProcessError(err)
	}

	// Выключаем автоперенос, по умолчанию 20
	pdf.SetAutoPageBreak(true, 9)
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report load testing for "+timeperiodstart.Format("02.01.2006 15:04")+"-"+timeperiodend.Format("15:04"), "1", 0, "CM", false, 0, "")
}

func ReportEnd() {
	// сжатие
	pdf.SetCompression(true)
	// запись отчета
	err := pdf.OutputFileAndClose(cfg.ReportPath + reportfilename + ".pdf")
	if err != nil {
		logs.ProcessError(err)
	}
}

func ReportProblemPDF() {
	logs.ProcessDebug("Start generate pdf - Problem ")
	saveX, saveY := pdf.GetXY()

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 15, 25, 120, 20} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Tag", "Type", "Name", "Description", "Threshold"}
	pdf.SetXY(saveX, saveY)
	pdf.SetY(pdf.GetY() + 8)

	pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	pdf.SetFont("Times", "B", 10)
	pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	pdf.Ln(-1)
	for i, item := range Problems {
		if item.Tag == "Jmeter" {
			continue
		}
		logs.ProcessDebug(item)
		pdf.SetFont("Times", "", 5)

		// Получаем данные для строки
		rowData := []string{item.Tag, item.Type, item.Name, item.Description, fmt.Sprint(item.Threshold)}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}

		g, _ := pdf.GetFontSize()
		rowHeight := float64(maxLines) * g

		// Проверяем, не нужно ли добавить новую страницу
		if pdf.GetY()+rowHeight+10 > float64(a4height) {
			pdf.AddPage()
			saveY = 10
			pdf.SetXY(saveX, saveY)
			// Повторяем заголовки на новой странице
			pdf.SetFont("Times", "", 5)
			pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
		}

		// Чередование цвета фона для строк
		pdf.SetFont("Times", "", 5)
		saveX, saveY = pdf.GetXY()

		fill := i%2 == 1
		if fill {
			pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.MultiCell(colWidth[0], MultiCellSize(g, maxLines, []byte(item.Tag), colWidth[0]), item.Tag, "1", "C", fill)

		pdf.SetXY(saveX+colWidth[0], saveY)
		pdf.MultiCell(colWidth[1], MultiCellSize(g, maxLines, []byte(item.Type), colWidth[1]), item.Type, "1", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		pdf.MultiCell(colWidth[2], MultiCellSize(g, maxLines, []byte(item.Name), colWidth[2]), item.Name, "1", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		pdf.MultiCell(colWidth[3], MultiCellSize(g, maxLines, []byte(item.Description), colWidth[3]), item.Description, "1", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2]+colWidth[3], saveY)
		pdf.MultiCell(colWidth[4], MultiCellSize(g, maxLines, []byte(fmt.Sprint(item.Threshold)), colWidth[4]), fmt.Sprint(item.Threshold), "1", "C", fill)

		// Переходим к следующей строке
		pdf.SetXY(saveX, saveY+rowHeight)
	}
}

func ReportProblemScenPDF() {
	logs.ProcessDebug("Start generate pdf - Problem scenario")
	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report problem Jmeter", "0", 0, "CM", false, 0, "")
	pdf.SetY(pdf.GetY() + 6)

	saveX, saveY := pdf.GetXY()

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 15, 35, 110, 20} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Tag", "Type", "Name", "Description", "Threshold"}
	pdf.SetXY(saveX, saveY)
	pdf.SetY(pdf.GetY() + 8)

	pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	pdf.SetFont("Times", "B", 10)
	pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	pdf.Ln(-1)
	for i, item := range Problems {
		if item.Tag != "Jmeter" {
			continue
		}
		logs.ProcessDebug(item)
		pdf.SetFont("Times", "", 5)

		// Получаем данные для строки
		rowData := []string{item.Tag, item.Type, item.Name, item.Description, fmt.Sprint(item.Threshold)}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}

		}

		g, _ := pdf.GetFontSize()
		rowHeight := float64(maxLines) * g

		// Проверяем, не нужно ли добавить новую страницу
		if pdf.GetY()+rowHeight+10 > float64(a4height) {
			pdf.AddPage()
			saveY = 10
			pdf.SetXY(saveX, saveY)
			// Повторяем заголовки на новой странице
			pdf.SetFont("Times", "", 5)
			pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
			pdf.Ln(-1)
		}

		// Чередование цвета фона для строк
		pdf.SetFont("Times", "", 5)
		saveX, saveY = pdf.GetXY()

		fill := i%2 == 1
		if fill {
			pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		pdf.MultiCell(colWidth[0], MultiCellSize(g, maxLines, []byte(item.Tag), colWidth[0]), item.Tag, "1", "C", fill)

		pdf.SetXY(saveX+colWidth[0], saveY)
		pdf.MultiCell(colWidth[1], MultiCellSize(g, maxLines, []byte(item.Type), colWidth[1]), item.Type, "1", "C", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		pdf.MultiCell(colWidth[2], MultiCellSize(g, maxLines, []byte(item.Name), colWidth[2]), item.Name, "1", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		pdf.MultiCell(colWidth[3], MultiCellSize(g, maxLines, []byte(item.Description), colWidth[3]), item.Description, "1", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2]+colWidth[3], saveY)
		pdf.MultiCell(colWidth[4], MultiCellSize(g, maxLines, []byte(fmt.Sprint(item.Threshold)), colWidth[4]), fmt.Sprint(item.Threshold), "1", "C", fill)

		// Переходим к следующей строке
		pdf.SetXY(saveX, saveY+rowHeight)
	}
}
func ReportInfluxPDF() {
	logs.ProcessDebug("Start generate pdf - Influx JMeter")
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
	logs.ProcessDebug("Start generate pdf - Influx Scenario Jmeter")
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
			pdf.SetFont("Times", "B", 6)
			pdf.CellFormat(195, 7, i.NameTest+":"+i.NameThread, "0", 0, "CM", false, 0, "")
			pdf.SetY(pdf.GetY() + 6)

			scenariopdf := make(map[string][]reportdata.YField)
			// Для красивого вывода в отчет
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
	logs.ProcessDebug("Start generate pdf - Grafana")
	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)

	var saveX, saveY, tmpheight float64
	saveX = 10
	tmpheight = 0
	saveY = 6

	for _, i := range LTGrafs {

		//сохранение переменных для второго и послед рисунков
		file, err := os.Open("tmp/" + i.Name + ".png")
		if err != nil {
			logs.ProcessError(err)
			continue
		}

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
				logs.ProcessInfo(err)
			}
		}

	}
}

func ClickHouseReportPDF() {
	logs.ProcessDebug("Start generate pdf - ClickHouse")

	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report ClickHouse", "0", 0, "CM", false, 0, "")
	pdf.Ln(-1)
	saveX, saveY := pdf.GetXY()

	// добавить вычисление длины для таблиц

	for _, i := range LTClickHouse {

		if pdf.GetY()+float64(35) > a4height || saveY+35 > a4height {
			pdf.AddPage()
			saveY = 10
			pdf.SetXY(saveX, saveY)
		}

		pdf.SetXY(saveX, saveY)
		pdf.SetY(pdf.GetY() + 8)
		pdf.SetFont("Times", "B", 10)
		pdf.CellFormat(195, 7, i.Name, "0", 0, "CM", false, 0, "")
		pdf.Ln(-1)

		pdf.SetY(pdf.GetY() + 8)
		saveX, saveY = pdf.GetXY()

		// При создании новой таблицы проверяем, что она влазит на страницу
		// Добавить перенос строк
		// разобраться с отступом

		// Цвета для таблицы
		headerFillColor := [3]int{200, 200, 200} // серый для заголовков
		rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных

		pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
		pdf.SetFont("Times", "B", 5)
		pdf.SetTextColor(0, 0, 0)

		//_, g := pdf.GetFontSize()
		g := pdf.GetStringWidth("3")
		// Округление
		i.RoundToPrecision(4)
		// Вычисляем длину
		tmpint := i.Lens()
		// Расчет ширины колонок
		var sum int
		for key, value := range tmpint {
			if value > 15 {
				tmpint[key] = int(float64(value) * g)
				sum += int(float64(value) * g)
			} else {
				tmpint[key] = 15
				sum += 15
			}
		}

		cof := 190 / float64(sum)

		tmp := make(map[string]float64)
		for key, value := range tmpint {
			tmp[key] = float64(value) * cof
		}

		pdf.SetFont("Times", "B", 5)
		for _, header := range i.Meta {
			pdf.CellFormat(tmp[header.Name], 5, header.Name, "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)

		for fi, jj := range i.Data {

			// Рассчитываем максимальную высоту для строки
			maxLines := 1
			for _, jjj := range i.Meta {
				lines := SplitTextPdf(jj[jjj.Name].(string), tmp[jjj.Name])
				if len(lines) > maxLines {
					maxLines = len(lines)
				}
			}

			gs, _ := pdf.GetFontSize()
			rowHeight := float64(maxLines) * gs

			if pdf.GetY()+rowHeight+10 > float64(a4height) {
				pdf.AddPage()
				saveY = 10
			}

			fill := fi%2 == 1
			if fill {
				pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			pdf.SetFont("Times", "", 5)
			saveX, saveY = pdf.GetXY()
			saveXs := saveX

			for _, jjj := range i.Meta {
				pdf.SetXY(saveX, saveY)
				pdf.MultiCell(tmp[jjj.Name], MultiCellSize(gs, maxLines, []byte(jj[jjj.Name].(string)), tmp[jjj.Name]), jj[jjj.Name].(string), "T", "L", fill)
				saveX += tmp[jjj.Name]
			}
			pdf.SetXY(saveXs, saveY+rowHeight)
		}
		saveX, saveY = pdf.GetXY()
	}
}

func ReportIMPDF() {
	logs.ProcessDebug("Start generate pdf - FSM")
	var saveX, saveY float64

	pdf.AddPage()
	pdf.SetY(pdf.GetY() + 6)
	pdf.SetFont("Times", "B", 16)
	pdf.CellFormat(195, 7, "Report FSM", "0", 0, "CM", false, 0, "")
	saveX, saveY = pdf.GetXY()

	ln := pdf.PointConvert(6)

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 45, 80, 50} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Number", "Brief Description", "Description", "AVR Name"}
	pdf.SetXY(saveX, saveY)
	pdf.SetY(pdf.GetY() + 8)

	pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	pdf.SetFont("Times", "B", 10)
	pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	for i, item := range LTIM.Content {
		// Чередование цвета фона для строк
		pdf.Ln(-1)
		pdf.SetFont("Times", "", 5)

		fill := i%2 == 1
		if fill {
			pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			pdf.SetFillColor(255, 255, 255)
		}

		// Получаем данные для строки
		rowData := []string{item.Number, item.BriefDescription, item.Description, item.AvrName}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}

		}

		saveX, saveY = pdf.GetXY()
		rowHeight := float64(maxLines-1) * 5

		pdf.MultiCell(colWidth[0], 5, item.Number, "T", "C", fill)

		pdf.SetXY(saveX+colWidth[0], saveY)
		pdf.MultiCell(colWidth[1], 5, item.BriefDescription, "T", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		pdf.MultiCell(colWidth[2], 5, item.Description, "T", "L", fill)

		pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		pdf.MultiCell(colWidth[3], 5, item.AvrName, "T", "L", fill)

		// Переходим к следующей строке
		pdf.SetXY(saveX, saveY+rowHeight)

		// Проверяем, не нужно ли добавить новую страницу
		if pdf.GetY()+float64(25)+ln > a4height || saveY+25+ln > a4height {
			pdf.AddPage()
			saveY = 10
			saveX = 6
			// Повторяем заголовки на новой странице
			pdf.SetFont("Times", "", 5)
			pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
		}
	}

}

func MultiCellSize(fontsize float64, maxLines int, text []byte, colWidth float64) float64 {
	lines := SplitTextPdf(string(text), colWidth)
	height := float64(maxLines) * fontsize
	return height / float64(len(lines))
}

// IsUTF8Supported быстрая проверка поддержки UTF-8
func IsUTF8Supported() bool {
	// Критерии проверки:
	// 1. Ширина кириллических символов > 0
	// 2. Ширина специальных символов > 0
	// 3. Отношение ширины Unicode к ASCII разумное

	asciiWidth := pdf.GetStringWidth("Hello")
	unicodeWidth := pdf.GetStringWidth("Привет")
	specialWidth := pdf.GetStringWidth("©€")

	// Если Unicode символы имеют разумную ширину
	hasUnicodeSupport := unicodeWidth > 0 && unicodeWidth > asciiWidth*0.3
	hasSpecialSupport := specialWidth > 0

	return hasUnicodeSupport && hasSpecialSupport
}

func SplitTextPdf(text string, colWidth float64) []string {
	if pdf == nil {
		return []string{text} // Возвращаем исходный текст как одну строку
	}

	// Проверяем валидность параметров
	if text == "" || colWidth <= 0 {
		return []string{text}
	}

	var lines []string
	if IsUTF8Supported() {
		lines = pdf.SplitText(text, colWidth)
	} else {
		lin := pdf.SplitLines([]byte(text), colWidth)
		for _, j := range lin {
			lines = append(lines, string(j))
		}
	}
	return lines
}
