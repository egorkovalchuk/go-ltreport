package fonts

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

//go:embed times.ttf
var timesRegular []byte

//go:embed timesbd.ttf
var timesBold []byte

//go:embed timesi.ttf
var timesItalic []byte

//go:embed timesbi.ttf
var timesBoldItalic []byte

type FontManager struct {
	fontsLoaded bool
	fontDir     string
}

func NewFontManager() *FontManager {
	// Создаем временную папку для шрифтов
	tempDir := "./temp_fonts"
	os.MkdirAll(tempDir, 0755)

	return &FontManager{
		fontDir: tempDir,
	}
}

func (fm *FontManager) Cleanup() {
	// Очищаем временные файлы при завершении
	os.RemoveAll(fm.fontDir)
}

func (fm *FontManager) SetupFonts(pdf *gofpdf.Fpdf) error {
	if fm.fontsLoaded {
		return nil
	}

	fontData := map[string]struct {
		data  []byte
		style string
	}{
		"times.ttf":   {timesRegular, ""},
		"timesbd.ttf": {timesBold, "B"},
		"timesi.ttf":  {timesItalic, "I"},
		"timesbi.ttf": {timesBoldItalic, "BI"},
	}

	for filename, font := range fontData {
		fontPath := filepath.Join(fm.fontDir, filename)

		// Сохраняем шрифт во временный файл
		if err := os.WriteFile(fontPath, font.data, 0644); err != nil {
			return fmt.Errorf("Error saving font %s: %w", filename, err)
		}

		// Добавляем шрифт в PDF
		pdf.AddUTF8Font("Times", font.style, fontPath)

		fm.fontsLoaded = true
	}
	return nil
}
