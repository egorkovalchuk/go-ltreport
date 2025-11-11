package reportdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// CalculatePercentile вычисляет заданный персентиль для набора значений
func CalculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return math.NaN()
	}

	// Сортируем значения
	sort.Float64s(values)

	// Вычисляем индекс персентиля
	index := (percentile / 100) * float64(len(values)-1)

	// Если индекс целый - возвращаем соответствующее значение
	if index == float64(int(index)) {
		return values[int(index)]
	}

	// Интерполируем между соседними значениями
	i := int(index)
	fraction := index - float64(i)
	return values[i] + fraction*(values[i+1]-values[i])
}

// HasTrailingSeparator проверяет наличие разделителя в конце пути
func HasTrailingSeparator(path string) bool {
	if path == "" {
		return false
	}
	return strings.HasSuffix(path, string(filepath.Separator)) ||
		strings.HasSuffix(path, "/") // для URL и универсальности
}

// EnsureTrailingSeparator добавляет разделитель в конце пути
func EnsureTrailingSeparator(path string) string {
	if path == "" {
		return string(filepath.Separator)
	}
	if !HasTrailingSeparator(path) {
		return path + string(filepath.Separator)
	}
	return path
}

// RemoveTrailingSeparator удаляет разделитель в конце пути
func RemoveTrailingSeparator(path string) string {
	if HasTrailingSeparator(path) {
		// Удаляем все trailing separators
		return strings.TrimRight(path, string(filepath.Separator))
	}
	return path
}

func RoundToPrecision(value float64, precision int) float64 {
	if precision < 0 {
		return value
	}

	ratio := math.Pow(10, float64(precision))

	// Обработка отрицательных чисел
	if value < 0 {
		return -math.Round(-value*ratio) / ratio
	}

	return math.Round(value*ratio) / ratio
}

// ConvertEncoding преобразует текст из указанной кодировки в UTF-8
func ConvertEncoding(text []byte) (string, error) {

	code := DetectEncodingFromBytes(text)
	encodings := map[string]*charmap.Charmap{
		"Windows-1251": charmap.Windows1251,
		"KOI8-R":       charmap.KOI8R,
		"ISO-8859-5":   charmap.ISO8859_5,
	}

	if code != "UTF-8" {
		if code != "Unknown" {
			decoder := encodings[code].NewDecoder()
			reader := transform.NewReader(bytes.NewReader(text), decoder)
			result, err := io.ReadAll(reader)
			if err != nil {
				return "", err
			}
			return string(result), nil
		} else {
			return "", err
		}
	} else {
		return string(text), nil
	}
}

func DetectEncodingFromBytes(data []byte) string {
	// Проверка на UTF-8 с BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "UTF-8 with BOM"
	}

	// Проверка на UTF-16 LE
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return "UTF-16 LE"
	}

	// Проверка на UTF-16 BE
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return "UTF-16 BE"
	}

	// Проверка на UTF-8 без BOM
	if utf8.Valid(data) {
		return "UTF-8"
	}

	// Попробуем определить русские кодировки
	if isLikelyWindows1251(data) {
		return "Windows-1251"
	}

	if isLikelyKOI8R(data) {
		return "KOI8-R"
	}

	if isLikelyISO8859_5(data) {
		return "ISO-8859-5"
	}

	return "Unknown"
}

// Вспомогательные функции для определения русских кодировок
func isLikelyWindows1251(data []byte) bool {
	// Windows-1251: характерные байты для русских букв
	win1251Ranges := []struct{ min, max byte }{
		{0xC0, 0xFF}, // Русские буквы
		{0x80, 0xBF}, // Дополнительные символы
	}

	return checkByteRanges(data, win1251Ranges)
}

func isLikelyKOI8R(data []byte) bool {
	// KOI8-R: характерные байты для русских букв
	koi8rRanges := []struct{ min, max byte }{
		{0x80, 0xFF}, // Русские буквы (в верхней половине)
		{0xC0, 0xDF}, // Прописные русские
		{0xE0, 0xFF}, // Строчные русские
	}

	return checkByteRanges(data, koi8rRanges)
}

func isLikelyISO8859_5(data []byte) bool {
	// ISO-8859-5: русские буквы начинаются с 0xB0
	iso8859_5Ranges := []struct{ min, max byte }{
		{0xB0, 0xCF}, // Русские буквы часть 1
		{0xD0, 0xFF}, // Русские буквы часть 2
	}

	return checkByteRanges(data, iso8859_5Ranges)
}

func checkByteRanges(data []byte, ranges []struct{ min, max byte }) bool {
	count := 0
	total := 0

	for _, b := range data {
		if isEnglishLetterOrBasicChar(b) {
			continue
		}
		total++
		for _, r := range ranges {
			if b >= r.min && b <= r.max {
				count++
				break
			}
		}
	}

	// Если более 10% байт попадают в характерные диапазоны
	return total > 0 && float64(count)/float64(total) > 0.1
}

func isEnglishLetterOrBasicChar(b byte) bool {
	// Английские буквы a-z, A-Z
	if (b >= 0x41 && b <= 0x5A) || (b >= 0x61 && b <= 0x7A) {
		return true
	}
	// Цифры 0-9
	if b >= 0x30 && b <= 0x39 {
		return true
	}
	// Базовые символы: пробел, пунктуация и т.д.
	basicSymbols := []byte{
		0x20, // space
		0x09, // tab
		0x0A, // new line
		0x0D, // carriage return
		0x21, // !
		0x22, // "
		0x23, // #
		0x24, // $
		0x25, // %
		0x26, // &
		0x27, // '
		0x28, // (
		0x29, // )
		0x2A, // *
		0x2B, // +
		0x2C, // ,
		0x2D, // -
		0x2E, // .
		0x2F, // /
		0x3A, // :
		0x3B, // ;
		0x3C, // <
		0x3D, // =
		0x3E, // >
		0x3F, // ?
		0x40, // @
		0x5B, // [
		0x5C, // \
		0x5D, // ]
		0x5E, // ^
		0x5F, // _
		0x60, // `
		0x7B, // {
		0x7C, // |
		0x7D, // }
		0x7E, // ~
	}
	for _, symbol := range basicSymbols {
		if b == symbol {
			return true
		}
	}
	return false
}

func MaxInt(x int, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func MaxInt64(x int64, y int64) int64 {
	if x > y {
		return x
	} else {
		return y
	}
}

func MinInt64(x int64, y int64) int64 {
	if x < y {
		return x
	} else {
		return y
	}
}

func ConvIntefaceFloat64(p interface{}) (float64, bool) {

	switch v := p.(type) {
	case json.Number:
		tmp, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return tmp, true
	case string:
		tmp, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return tmp, true
	case int:
		return float64(v), false
	case float64:
		return v, true
	case bool:
		return 0, false
	default:
		return 0, false
	}
}

func ConvIntefaceInt64(p interface{}) (int64, bool) {
	switch v := p.(type) {
	case json.Number:
		tmp, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return tmp, true
	case string:
		tmp, err := strconv.Atoi(v)
		if err != nil {
			return 0, false
		}
		return int64(tmp), true
	case int:
		return int64(v), false
	case float64:
		return int64(v), true
	case bool:
		return 0, false
	default:
		return 0, false
	}
}

func CheckStatusCode(StatusCode int, Status string) error {
	switch StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		return nil
	case http.StatusNoContent, http.StatusResetContent:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed")
	case http.StatusServiceUnavailable:
		return fmt.Errorf("service is not available: %s", Status)
	case http.StatusInternalServerError:
		return fmt.Errorf("internal server error: %s", Status)
	case http.StatusConflict:
		return fmt.Errorf("conflict: %s", Status)
	default:
		return fmt.Errorf("unknown response status: %s", Status)
	}
}

func BeginningOfDay() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, time.Local)
}

// EndOfDay end of day
func EndOfDay() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 17, 59, 59, int(time.Second-time.Nanosecond), time.Local)
}

func BeginningOfHour() time.Time {
	t := time.Now()
	now := t.Hour()
	return time.Date(t.Year(), t.Month(), t.Day(), now-1, 0, 0, 0, time.Local)
}

// EndOfDay end of day
func EndOfHour() time.Time {
	t := time.Now()
	now := t.Hour()
	return time.Date(t.Year(), t.Month(), t.Day(), now-1, 59, 59, int(time.Second-time.Nanosecond), time.Local)
}
