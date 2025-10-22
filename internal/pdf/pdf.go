package pdf

import (
	"fmt"
	"os"

	"time"

	"github.com/egorkovalchuk/go-ltreport/internal/confluence"
	"github.com/egorkovalchuk/go-ltreport/internal/fonts"
	"github.com/egorkovalchuk/go-ltreport/internal/hpsm"
	"github.com/egorkovalchuk/go-ltreport/internal/logger"
	"github.com/egorkovalchuk/go-ltreport/internal/reportdata"

	"github.com/jung-kurt/gofpdf"
)

const (
	a4height = 297
	a4width  = 210
)

var (
	// launch links
	allurelink string
)

type PDFWriter struct {
	logs       *logger.LogWriter
	pdf        *gofpdf.Fpdf
	start, end time.Time
	filename   string
	path       string
	confcfg    reportdata.ConfluenceType
}

func NewPDFWhriter(filename, path string, start, end time.Time, confcfg reportdata.ConfluenceType, logFunc *logger.LogWriter) *PDFWriter {
	return &PDFWriter{
		logs:     logFunc,
		filename: filename + ".pdf",
		start:    start,
		end:      end,
		path:     path,
		confcfg:  confcfg,
	}
}

func (p *PDFWriter) ReportPDFInit() {

	fontManager := fonts.NewFontManager()
	defer fontManager.Cleanup()

	p.logs.ProcessInfo("File name " + p.filename)
	// Инициализация pdf
	p.pdf = gofpdf.New("P", "mm", "A4", "")
	p.pdf.AddPage()

	// Настраиваем шрифты
	if err := fontManager.SetupFonts(p.pdf); err != nil {
		p.logs.ProcessError(err)
	}

	// Выключаем автоперенос, по умолчанию 20
	p.pdf.SetAutoPageBreak(true, 9)
	p.pdf.SetY(p.pdf.GetY() + 6)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report load testing for "+p.start.Format("02.01.2006 15:04")+"-"+p.end.Format("15:04"), "1", 0, "CM", false, 0, "")
}

func (p *PDFWriter) ReportEnd() {
	// сжатие
	p.pdf.SetCompression(true)
	// запись отчета
	err := p.pdf.OutputFileAndClose(p.path + p.filename)
	if err != nil {
		p.logs.ProcessError(err)
	}
}

func (p *PDFWriter) ReportProblemPDF(Problems []reportdata.LTError) {
	p.logs.ProcessDebug("Start generate pdf - Problem ")
	saveX, saveY := p.pdf.GetXY()

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 15, 25, 120, 20} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Tag", "Type", "Name", "Description", "Threshold"}
	p.pdf.SetXY(saveX, saveY)
	p.pdf.SetY(p.pdf.GetY() + 8)

	p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	p.pdf.SetFont("Times", "B", 10)
	p.pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		p.pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	p.pdf.Ln(-1)
	for i, item := range Problems {
		if item.Tag == "Jmeter" {
			continue
		}
		p.logs.ProcessDebug(item)
		p.pdf.SetFont("Times", "", 5)

		// Получаем данные для строки
		rowData := []string{item.Tag, item.Type, item.Name, item.Description, fmt.Sprint(item.Threshold)}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := p.SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}
		}

		g, _ := p.pdf.GetFontSize()
		rowHeight := float64(maxLines) * g

		// Проверяем, не нужно ли добавить новую страницу
		if p.pdf.GetY()+rowHeight+10 > float64(a4height) {
			p.pdf.AddPage()
			saveY = 10
			p.pdf.SetXY(saveX, saveY)
			// Повторяем заголовки на новой странице
			p.pdf.SetFont("Times", "", 5)
			p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				p.pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
			p.pdf.Ln(-1)
		}

		// Чередование цвета фона для строк
		p.pdf.SetFont("Times", "", 5)
		saveX, saveY = p.pdf.GetXY()

		fill := i%2 == 1
		if fill {
			p.pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			p.pdf.SetFillColor(255, 255, 255)
		}

		p.pdf.MultiCell(colWidth[0], p.MultiCellSize(g, maxLines, []byte(item.Tag), colWidth[0]), item.Tag, "1", "C", fill)

		p.pdf.SetXY(saveX+colWidth[0], saveY)
		p.pdf.MultiCell(colWidth[1], p.MultiCellSize(g, maxLines, []byte(item.Type), colWidth[1]), item.Type, "1", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		p.pdf.MultiCell(colWidth[2], p.MultiCellSize(g, maxLines, []byte(item.Name), colWidth[2]), item.Name, "1", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		p.pdf.MultiCell(colWidth[3], p.MultiCellSize(g, maxLines, []byte(item.Description), colWidth[3]), item.Description, "1", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2]+colWidth[3], saveY)
		p.pdf.MultiCell(colWidth[4], p.MultiCellSize(g, maxLines, []byte(fmt.Sprint(item.Threshold)), colWidth[4]), fmt.Sprint(item.Threshold), "1", "C", fill)

		// Переходим к следующей строке
		p.pdf.SetXY(saveX, saveY+rowHeight)
	}
}

func (p *PDFWriter) ReportProblemScenPDF(Problems []reportdata.LTError) {
	p.logs.ProcessDebug("Start generate pdf - Problem scenario")
	p.pdf.AddPage()
	p.pdf.SetY(p.pdf.GetY() + 6)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report problem Jmeter", "0", 0, "CM", false, 0, "")
	p.pdf.SetY(p.pdf.GetY() + 6)

	saveX, saveY := p.pdf.GetXY()

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 15, 35, 110, 20} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Tag", "Type", "Name", "Description", "Threshold"}
	p.pdf.SetXY(saveX, saveY)
	p.pdf.SetY(p.pdf.GetY() + 8)

	p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	p.pdf.SetFont("Times", "B", 10)
	p.pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		p.pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	p.pdf.Ln(-1)
	for i, item := range Problems {
		if item.Tag != "Jmeter" {
			continue
		}
		p.logs.ProcessDebug(item)
		p.pdf.SetFont("Times", "", 5)

		// Получаем данные для строки
		rowData := []string{item.Tag, item.Type, item.Name, item.Description, fmt.Sprint(item.Threshold)}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := p.SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}

		}

		g, _ := p.pdf.GetFontSize()
		rowHeight := float64(maxLines) * g

		// Проверяем, не нужно ли добавить новую страницу
		if p.pdf.GetY()+rowHeight+10 > float64(a4height) {
			p.pdf.AddPage()
			saveY = 10
			p.pdf.SetXY(saveX, saveY)
			// Повторяем заголовки на новой странице
			p.pdf.SetFont("Times", "", 5)
			p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				p.pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
			p.pdf.Ln(-1)
		}

		// Чередование цвета фона для строк
		p.pdf.SetFont("Times", "", 5)
		saveX, saveY = p.pdf.GetXY()

		fill := i%2 == 1
		if fill {
			p.pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			p.pdf.SetFillColor(255, 255, 255)
		}

		p.pdf.MultiCell(colWidth[0], p.MultiCellSize(g, maxLines, []byte(item.Tag), colWidth[0]), item.Tag, "1", "C", fill)

		p.pdf.SetXY(saveX+colWidth[0], saveY)
		p.pdf.MultiCell(colWidth[1], p.MultiCellSize(g, maxLines, []byte(item.Type), colWidth[1]), item.Type, "1", "C", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		p.pdf.MultiCell(colWidth[2], p.MultiCellSize(g, maxLines, []byte(item.Name), colWidth[2]), item.Name, "1", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		p.pdf.MultiCell(colWidth[3], p.MultiCellSize(g, maxLines, []byte(item.Description), colWidth[3]), item.Description, "1", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2]+colWidth[3], saveY)
		p.pdf.MultiCell(colWidth[4], p.MultiCellSize(g, maxLines, []byte(fmt.Sprint(item.Threshold)), colWidth[4]), fmt.Sprint(item.Threshold), "1", "C", fill)

		// Переходим к следующей строке
		p.pdf.SetXY(saveX, saveY+rowHeight)
	}
}
func (p *PDFWriter) ReportInfluxPDF(LTTest_dinamic []reportdata.LTTestDinamic) {
	p.logs.ProcessDebug("Start generate pdf - Influx JMeter")
	p.pdf.AddPage()

	p.pdf.SetY(p.pdf.GetY() + 10)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report test", "0", 0, "CM", false, 0, "")
	p.pdf.SetY(p.pdf.GetY() + 7)

	for _, i := range LTTest_dinamic {
		p.pdf.SetFont("Times", "B", 10)
		p.pdf.CellFormat(195, 7, i.NameTest, "0", 0, "CM", false, 0, "")

		p.pdf.SetFont("Times", "", 8)
		p.pdf.SetY(p.pdf.GetY() + 7)

		num := 1
		for _, ii := range i.Field {

			p.pdf.CellFormat(65, 7, fmt.Sprintf(ii.Description, int(ii.Value)), "1", 0, "LM", false, 0, "")

			if num%3 == 0 {
				p.pdf.SetY(p.pdf.GetY() + 7)
			}
			num = num + 1
		}

	}

}

func (p *PDFWriter) ReportInfluxScrnPDF(LTScen_dimanict map[string]map[string]reportdata.ScenarioDinamic) {
	p.logs.ProcessDebug("Start generate pdf - Influx Scenario Jmeter")
	p.pdf.AddPage()
	p.pdf.SetY(p.pdf.GetY() + 10)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report scenario", "0", 0, "CM", false, 0, "")

	var keys []string
	for k := range LTScen_dimanict {
		keys = append(keys, k)
	}

	for _, k := range keys {

		p.pdf.SetY(p.pdf.GetY() + 6)
		p.pdf.SetFont("Times", "B", 8)
		p.pdf.CellFormat(195, 7, k, "0", 0, "CM", false, 0, "")

		for _, i := range LTScen_dimanict[k] {
			p.pdf.SetFont("Times", "B", 6)
			p.pdf.CellFormat(195, 7, i.NameTest+":"+i.NameThread, "0", 0, "CM", false, 0, "")
			p.pdf.SetY(p.pdf.GetY() + 6)

			scenariopdf := make(map[string][]reportdata.YField)
			// Для красивого вывода в отчет
			for _, ii := range i.Field {
				scenariopdf[ii.Statut] = append(scenariopdf[ii.Statut], ii)
			}

			if _, ok := scenariopdf["all"]; ok {
				p.pdf.SetFont("Times", "", 4)
				p.pdf.CellFormat(20, 4, "All ", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["all"] {
					p.pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						p.pdf.SetY(p.pdf.GetY() + 4)
					}
					num++
				}
			}

			if _, ok := scenariopdf["ok"]; ok {

				p.pdf.CellFormat(20, 4, "Status OK", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["ok"] {
					p.pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						p.pdf.SetY(p.pdf.GetY() + 4)
					}
					num++
				}
			}

			if _, ok := scenariopdf["ko"]; ok {

				p.pdf.CellFormat(20, 4, "Error ", "0", 0, "LM", false, 0, "")

				num := 1
				for _, ipdf := range scenariopdf["ko"] {
					p.pdf.CellFormat(40, 4, fmt.Sprintf(ipdf.Description, int(ipdf.Value)), "1", 0, "LM", false, 0, "")
					if num%4 == 0 {
						p.pdf.SetY(p.pdf.GetY() + 4)
					}
					num++
				}
			}

		}
	}

}

func (p *PDFWriter) GrafanaReportPDF(LTGrafs []reportdata.LTGrag) {
	p.logs.ProcessDebug("Start generate pdf - Grafana")
	p.pdf.AddPage()
	p.pdf.SetY(p.pdf.GetY() + 6)

	var saveX, saveY, tmpheight float64
	saveX = 10
	tmpheight = 0
	saveY = 6

	for _, i := range LTGrafs {

		//сохранение переменных для второго и послед рисунков
		file, err := os.Open("tmp/" + i.Name + ".png")
		if err != nil {
			p.logs.ProcessError(err)
			continue
		}

		ln := p.pdf.PointConvert(8)

		if saveX+float64(i.Size.Width) > a4width {
			p.pdf.SetXY(10, saveY+tmpheight+ln+2)
			saveX = 10
			saveY = saveY + tmpheight + ln + 2
		}

		if p.pdf.GetY()+ln+float64(i.Size.Height)+7 > a4height || saveY > a4height {
			p.pdf.AddPage()
			saveY = 6
		}

		p.pdf.SetFont("Times", "B", 8)
		p.pdf.SetXY(saveX, saveY)
		p.pdf.MultiCell(float64(i.Size.Width), ln, i.Name, "", "C", false)

		tp := p.pdf.ImageTypeFromMime(i.ContentType)

		_ = p.pdf.RegisterImageOptionsReader(i.Name, gofpdf.ImageOptions{ImageType: tp, ReadDpi: false, AllowNegativePosition: true}, file)
		if p.pdf.Ok() {

			p.pdf.Image(i.Name, saveX, saveY+4,
				float64(i.Size.Width), float64(i.Size.Height), false, "png", 0, i.UrlDash)
			saveX = saveX + float64(i.Size.Width) + 10

		}

		tmpheight = float64(i.Size.Height)

		defer file.Close()
		file.Close()
	}
}

func (p *PDFWriter) ClickHouseReportPDF(LTClickHouse []reportdata.ClickHouseJson) {
	p.logs.ProcessDebug("Start generate pdf - ClickHouse")

	p.pdf.AddPage()
	p.pdf.SetY(p.pdf.GetY() + 6)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report ClickHouse", "0", 0, "CM", false, 0, "")
	p.pdf.Ln(-1)
	saveX, saveY := p.pdf.GetXY()

	// добавить вычисление длины для таблиц

	for _, i := range LTClickHouse {

		if p.pdf.GetY()+float64(35) > a4height || saveY+35 > a4height {
			p.pdf.AddPage()
			saveY = 10
			p.pdf.SetXY(saveX, saveY)
		}

		p.pdf.SetXY(saveX, saveY)
		p.pdf.SetY(p.pdf.GetY() + 8)
		p.pdf.SetFont("Times", "B", 10)
		p.pdf.CellFormat(195, 7, i.Name, "0", 0, "CM", false, 0, "")
		p.pdf.Ln(-1)

		p.pdf.SetY(p.pdf.GetY() + 8)
		saveX, saveY = p.pdf.GetXY()

		// При создании новой таблицы проверяем, что она влазит на страницу
		// Добавить перенос строк
		// разобраться с отступом

		// Цвета для таблицы
		headerFillColor := [3]int{200, 200, 200} // серый для заголовков
		rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных

		p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
		p.pdf.SetFont("Times", "B", 5)
		p.pdf.SetTextColor(0, 0, 0)

		//_, g := p.pdf.GetFontSize()
		g := p.pdf.GetStringWidth("3")
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

		p.pdf.SetFont("Times", "B", 5)
		for _, header := range i.Meta {
			p.pdf.CellFormat(tmp[header.Name], 5, header.Name, "1", 0, "C", true, 0, "")
		}
		p.pdf.Ln(-1)

		for fi, jj := range i.Data {

			// Рассчитываем максимальную высоту для строки
			maxLines := 1
			for _, jjj := range i.Meta {
				lines := p.SplitTextPdf(jj[jjj.Name].(string), tmp[jjj.Name])
				if len(lines) > maxLines {
					maxLines = len(lines)
				}
			}

			gs, _ := p.pdf.GetFontSize()
			rowHeight := float64(maxLines) * gs

			if p.pdf.GetY()+rowHeight+10 > float64(a4height) {
				p.pdf.AddPage()
				saveY = 10
			}

			fill := fi%2 == 1
			if fill {
				p.pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
			} else {
				p.pdf.SetFillColor(255, 255, 255)
			}

			p.pdf.SetFont("Times", "", 5)
			saveX, saveY = p.pdf.GetXY()
			saveXs := saveX

			for _, jjj := range i.Meta {
				p.pdf.SetXY(saveX, saveY)
				p.pdf.MultiCell(tmp[jjj.Name], p.MultiCellSize(gs, maxLines, []byte(jj[jjj.Name].(string)), tmp[jjj.Name]), jj[jjj.Name].(string), "T", "L", fill)
				saveX += tmp[jjj.Name]
			}
			p.pdf.SetXY(saveXs, saveY+rowHeight)
		}
		saveX, saveY = p.pdf.GetXY()
	}
}

func (p *PDFWriter) ReportIMPDF(LTIM hpsm.Content) {
	p.logs.ProcessDebug("Start generate pdf - FSM")
	var saveX, saveY float64

	p.pdf.AddPage()
	p.pdf.SetY(p.pdf.GetY() + 6)
	p.pdf.SetFont("Times", "B", 16)
	p.pdf.CellFormat(195, 7, "Report FSM", "0", 0, "CM", false, 0, "")
	saveX, saveY = p.pdf.GetXY()

	ln := p.pdf.PointConvert(6)

	// Цвета для таблицы
	headerFillColor := [3]int{200, 200, 200} // серый для заголовков
	rowFillColor := [3]int{240, 240, 240}    // светлый серый для четных строк

	// Ширина колонок (сумма = 190 для A4 формата с полями)
	colWidth := []float64{15, 45, 80, 50} // number, avrName, briefDescription, description

	// Заголовки колонок
	headers := []string{"Number", "Brief Description", "Description", "AVR Name"}
	p.pdf.SetXY(saveX, saveY)
	p.pdf.SetY(p.pdf.GetY() + 8)

	p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
	p.pdf.SetFont("Times", "B", 10)
	p.pdf.SetTextColor(0, 0, 0)

	for i, header := range headers {
		p.pdf.CellFormat(colWidth[i], 5, header, "1", 0, "C", true, 0, "")
	}

	for i, item := range LTIM.Content {
		// Чередование цвета фона для строк
		p.pdf.Ln(-1)
		p.pdf.SetFont("Times", "", 5)

		fill := i%2 == 1
		if fill {
			p.pdf.SetFillColor(rowFillColor[0], rowFillColor[1], rowFillColor[2])
		} else {
			p.pdf.SetFillColor(255, 255, 255)
		}

		// Получаем данные для строки
		rowData := []string{item.Number, item.BriefDescription, item.Description, item.AvrName}

		// Рассчитываем максимальную высоту для строки
		maxLines := 1
		for j, text := range rowData {
			lines := p.SplitTextPdf(text, colWidth[j])
			if len(lines) > maxLines {
				maxLines = len(lines)
			}

		}

		saveX, saveY = p.pdf.GetXY()
		rowHeight := float64(maxLines-1) * 5

		p.pdf.MultiCell(colWidth[0], 5, item.Number, "T", "C", fill)

		p.pdf.SetXY(saveX+colWidth[0], saveY)
		p.pdf.MultiCell(colWidth[1], 5, item.BriefDescription, "T", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1], saveY)
		p.pdf.MultiCell(colWidth[2], 5, item.Description, "T", "L", fill)

		p.pdf.SetXY(saveX+colWidth[0]+colWidth[1]+colWidth[2], saveY)
		p.pdf.MultiCell(colWidth[3], 5, item.AvrName, "T", "L", fill)

		// Переходим к следующей строке
		p.pdf.SetXY(saveX, saveY+rowHeight)

		// Проверяем, не нужно ли добавить новую страницу
		if p.pdf.GetY()+float64(25)+ln > a4height || saveY+25+ln > a4height {
			p.pdf.AddPage()
			saveY = 10
			saveX = 6
			// Повторяем заголовки на новой странице
			p.pdf.SetFont("Times", "", 5)
			p.pdf.SetFillColor(headerFillColor[0], headerFillColor[1], headerFillColor[2])
			for i, header := range headers {
				p.pdf.CellFormat(colWidth[i], 7, header, "1", 0, "C", true, 0, "")
			}
		}
	}

}

func (p *PDFWriter) MultiCellSize(fontsize float64, maxLines int, text []byte, colWidth float64) float64 {
	lines := p.SplitTextPdf(string(text), colWidth)
	height := float64(maxLines) * fontsize
	return height / float64(len(lines))
}

// IsUTF8Supported быстрая проверка поддержки UTF-8
func (p *PDFWriter) IsUTF8Supported() bool {
	// Критерии проверки:
	// 1. Ширина кириллических символов > 0
	// 2. Ширина специальных символов > 0
	// 3. Отношение ширины Unicode к ASCII разумное

	asciiWidth := p.pdf.GetStringWidth("Hello")
	unicodeWidth := p.pdf.GetStringWidth("Привет")
	specialWidth := p.pdf.GetStringWidth("©€")

	// Если Unicode символы имеют разумную ширину
	hasUnicodeSupport := unicodeWidth > 0 && unicodeWidth > asciiWidth*0.3
	hasSpecialSupport := specialWidth > 0

	return hasUnicodeSupport && hasSpecialSupport
}

func (p *PDFWriter) SplitTextPdf(text string, colWidth float64) []string {
	if p.pdf == nil {
		return []string{text} // Возвращаем исходный текст как одну строку
	}

	// Проверяем валидность параметров
	if text == "" || colWidth <= 0 {
		return []string{text}
	}

	var lines []string
	if p.IsUTF8Supported() {
		lines = p.pdf.SplitText(text, colWidth)
	} else {
		lin := p.pdf.SplitLines([]byte(text), colWidth)
		for _, j := range lin {
			lines = append(lines, string(j))
		}
	}
	return lines
}

// Загрузка в Confluence
func (p *PDFWriter) ReportDownload() string {
	// пробрасываем дебаг
	confluence.DebugFlag = true

	// Инициализация работы с конфленсом
	// перенеммные потом вынески в конфиг
	confl, err := confluence.NewAPI(p.confcfg.ReportConfluenceURL, p.confcfg.ReportConfluenceLogin, p.confcfg.ReportConfluencePass, p.confcfg.ReportConfluenceToken, p.confcfg.ReportConfluenceProxy)
	if err != nil {
		p.logs.ProcessError("Error connection to confluence")
		p.logs.ProcessError(err)
		return allurelink
	}
	// Получение описание базовой страницы
	p.logs.ProcessDebug("GetContent")
	JsonCont, err := confl.GetContent(p.confcfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: p.confcfg.ReportConfluenceSpace, Expand: []string{"children.page"}}, p.logs.ProcessLog)
	if err != nil {
		p.logs.ProcessError(err)
		p.logs.ProcessError(JsonCont)
		return allurelink
	}
	p.logs.ProcessDebug("GetContentChildPage")
	JsonConC, err := confl.GetContentChildPage(p.confcfg.ReportConfluenceId, confluence.ContentQuery{SpaceKey: p.confcfg.ReportConfluenceSpace, Limit: 250, Expand: []string{"children.page"}}, p.logs.ProcessLog)
	if err != nil {
		p.logs.ProcessError(err)
		p.logs.ProcessError(JsonConC)
		return allurelink
	}

	currentTime := time.Now()
	reportname := "Report" + currentTime.Format("20060102")

	var IdChild string

	// поиск по детям GetContentChildPage
	for _, i := range JsonConC.Results {
		if reportname == i.Title {
			IdChild = i.ID
			p.logs.ProcessDebug(i.Title + ", id=" + IdChild)
		} else {
			IdChild = ""
		}
	}

	if IdChild == "" {
		p.logs.ProcessDebug("Create child page " + reportname)
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
				Key: p.confcfg.ReportConfluenceSpace,
			},
		}

		JsonContC, err := confl.CreateContent(&data, p.logs.ProcessLog)
		if err != nil {
			p.logs.ProcessError(err)
			return allurelink
		}
		IdChild = JsonContC.ID

		file, err := os.OpenFile(p.path+p.filename, os.O_RDONLY, 0666)
		if err != nil {
			p.logs.ProcessError(err)
		}

		p.logs.ProcessDebug("Upload current attachments " + p.filename)
		arsp, err := confl.UploadAttachment(IdChild, p.filename, file, p.logs.ProcessLog)
		if err != nil {
			p.logs.ProcessError(err)
			p.logs.ProcessDebug(arsp)
		}
		defer file.Close()
		allurelink = arsp.Results[0].Links.Webui
	} else {
		p.logs.ProcessDebug("Load current attachments")
		arsp, err := confl.GetAttachments(IdChild, p.logs.ProcessLog)
		if err != nil {
			p.logs.ProcessError(err)
			p.logs.ProcessDebug(arsp)
		}

		chck := false
		attachid := "0"
		for _, j := range arsp.Results {
			if j.Title == p.filename {
				attachid = j.ID
				chck = true
			}
		}

		file, err := os.OpenFile(p.path+p.filename, os.O_RDONLY, 0666)
		if err != nil {
			p.logs.ProcessInfo(err)
		}

		if chck {
			p.logs.ProcessDebug("Update current attachments " + p.filename)
			arsp, err := confl.UpdateAttachment(IdChild, p.filename, attachid, file, p.logs.ProcessLog)
			if err != nil {
				p.logs.ProcessInfo(err)
				p.logs.ProcessDebug(arsp)
			}
			allurelink = arsp.Links.Webui
		} else {
			p.logs.ProcessDebug("Upload current attachments " + p.filename)
			arsp, err := confl.UploadAttachment(IdChild, p.filename, file, p.logs.ProcessLog)
			if err != nil {
				p.logs.ProcessInfo(err)
				p.logs.ProcessDebug(arsp)
			}
			allurelink = arsp.Results[0].Links.Webui
		}
		defer file.Close()
	}

	return allurelink
}
