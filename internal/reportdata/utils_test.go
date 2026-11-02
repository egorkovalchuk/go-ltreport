package reportdata

import (
	"math"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculatePercentile_Empty(t *testing.T) {
	values := []float64{}
	result := CalculatePercentile(values, 99)
	if !math.IsNaN(result) {
		t.Errorf("Expected NaN for empty input, got %f", result)
	}
}

func TestCalculatePercentile_SingleValue(t *testing.T) {
	values := []float64{42.0}
	result := CalculatePercentile(values, 99)
	if result != 42.0 {
		t.Errorf("Expected 42.0, got %f", result)
	}
}

func TestCalculatePercentile_50th(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	result := CalculatePercentile(values, 50)
	if result != 3.0 {
		t.Errorf("Expected median 3.0, got %f", result)
	}
}

func TestCalculatePercentile_0th(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	result := CalculatePercentile(values, 0)
	if result != 1.0 {
		t.Errorf("Expected min 1.0, got %f", result)
	}
}

func TestCalculatePercentile_100th(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5}
	result := CalculatePercentile(values, 100)
	if result != 5.0 {
		t.Errorf("Expected max 5.0, got %f", result)
	}
}

func TestCalculatePercentile_99th(t *testing.T) {
	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := CalculatePercentile(values, 99)
	expected := 9.0 + 0.91*(10.0-9.0)
	if result != expected {
		t.Errorf("Expected 9.91, got %f", result)
	}
}

func TestCalculatePercentile_NegativeValues(t *testing.T) {
	values := []float64{-10, -5, 0, 5, 10}
	result := CalculatePercentile(values, 50)
	if result != 0.0 {
		t.Errorf("Expected 0.0, got %f", result)
	}
}

func TestCalculatePercentile_UnsortedInput(t *testing.T) {
	values := []float64{5, 1, 3, 2, 4}
	result := CalculatePercentile(values, 50)
	if result != 3.0 {
		t.Errorf("Expected 3.0 from unsorted input, got %f", result)
	}
}

func TestHasTrailingSeparator_Empty(t *testing.T) {
	if HasTrailingSeparator("") {
		t.Error("Expected false for empty string")
	}
}

func TestHasTrailingSeparator_WithSeparator(t *testing.T) {
	if !HasTrailingSeparator("/home/") {
		t.Error("Expected true for path with separator")
	}
}

func TestHasTrailingSeparator_NoSeparator(t *testing.T) {
	if HasTrailingSeparator("/home/user") {
		t.Error("Expected false for path without trailing separator")
	}
}

func TestHasTrailingSeparator_URLSlash(t *testing.T) {
	if !HasTrailingSeparator("https://example.com/") {
		t.Error("Expected true for URL with trailing slash")
	}
}

func TestEnsureTrailingSeparator_AddsSeparator(t *testing.T) {
	result := EnsureTrailingSeparator("/home/user")
	if result != "/home/user/" {
		t.Errorf("Expected '/home/user/', got '%s'", result)
	}
}

func TestEnsureTrailingSeparator_NoAddNeeded(t *testing.T) {
	result := EnsureTrailingSeparator("/home/user/")
	if result != "/home/user/" {
		t.Errorf("Expected '/home/user/', got '%s'", result)
	}
}

func TestEnsureTrailingSeparator_Empty(t *testing.T) {
	result := EnsureTrailingSeparator("")
	if result != string(filepath.Separator) {
		t.Errorf("Expected filepath.Separator for empty string, got '%s'", result)
	}
}

func TestRemoveTrailingSeparator_Removes(t *testing.T) {
	result := RemoveTrailingSeparator("/home/user/")
	if result != "/home/user" {
		t.Errorf("Expected '/home/user', got '%s'", result)
	}
}

func TestRemoveTrailingSeparator_NoRemoval(t *testing.T) {
	result := RemoveTrailingSeparator("/home/user")
	if result != "/home/user" {
		t.Errorf("Expected '/home/user', got '%s'", result)
	}
}

func TestRoundToPrecision_Positive(t *testing.T) {
	result := RoundToPrecision(3.14159, 2)
	expected := 3.14
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRoundToPrecision_Negative(t *testing.T) {
	result := RoundToPrecision(-3.14159, 2)
	expected := -3.14
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestRoundToPrecision_NegativePrecision(t *testing.T) {
	result := RoundToPrecision(3.14159, -1)
	expected := 3.14159
	if result != expected {
		t.Errorf("Expected original value for negative precision, got %f", result)
	}
}

func TestRoundToPrecision_ZeroPrecision(t *testing.T) {
	result := RoundToPrecision(3.7, 0)
	expected := 4.0
	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{4, 4, 4},
		{-1, -2, -1},
	}
	for _, tc := range tests {
		result := MaxInt(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("MaxInt(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestMaxInt64(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 2},
		{5, 3, 5},
		{4, 4, 4},
		{-1, -2, -1},
	}
	for _, tc := range tests {
		result := MaxInt64(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("MaxInt64(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestMinInt64(t *testing.T) {
	tests := []struct {
		a, b, expected int64
	}{
		{1, 2, 1},
		{5, 3, 3},
		{4, 4, 4},
		{-1, -2, -2},
	}
	for _, tc := range tests {
		result := MinInt64(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("MinInt64(%d, %d) = %d, want %d", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestConvIntefaceFloat64_Float64(t *testing.T) {
	val := float64(42.5)
	result, ok := ConvIntefaceFloat64(val)
	if !ok || result != 42.5 {
		t.Errorf("Expected 42.5, true, got %f, %v", result, ok)
	}
}

func TestConvIntefaceFloat64_Int(t *testing.T) {
	val := int(10)
	result, ok := ConvIntefaceFloat64(val)
	if !ok || result != float64(10) {
		t.Errorf("Expected 10.0, true, got %f, %v", result, ok)
	}
}

func TestConvIntefaceFloat64_String(t *testing.T) {
	val := "3.14"
	result, ok := ConvIntefaceFloat64(val)
	if !ok || result != 3.14 {
		t.Errorf("Expected 3.14, true, got %f, %v", result, ok)
	}
}

func TestConvIntefaceFloat64_InvalidString(t *testing.T) {
	_, ok := ConvIntefaceFloat64("abc")
	if ok {
		t.Error("Expected false for invalid string")
	}
}

func TestConvIntefaceFloat64_Nil(t *testing.T) {
	_, ok := ConvIntefaceFloat64(nil)
	if ok {
		t.Error("Expected false for nil")
	}
}

func TestConvIntefaceInt64_Int64(t *testing.T) {
	val := 42
	result, ok := ConvIntefaceInt64(val)
	if !ok || result != int64(42) {
		t.Errorf("Expected 42, true, got %d, %v", result, ok)
	}
}

func TestConvIntefaceInt64_Float64(t *testing.T) {
	val := float64(42.0)
	result, ok := ConvIntefaceInt64(val)
	if !ok || result != 42 {
		t.Errorf("Expected 42, true, got %d, %v", result, ok)
	}
}

func TestConvIntefaceInt64_String(t *testing.T) {
	val := "123"
	result, ok := ConvIntefaceInt64(val)
	if !ok || result != 123 {
		t.Errorf("Expected 123, true, got %d, %v", result, ok)
	}
}

func TestDetectEncodingFromBytes_UTF8(t *testing.T) {
	data := []byte("Hello World")
	encoding := DetectEncodingFromBytes(data)
	if encoding != "UTF-8" {
		t.Errorf("Expected UTF-8, got %s", encoding)
	}
}

func TestDetectEncodingFromBytes_UTF8BOM(t *testing.T) {
	data := []byte{0xEF, 0xBB, 0xBF, 72, 101, 108, 108, 111}
	encoding := DetectEncodingFromBytes(data)
	if encoding != "UTF-8 with BOM" {
		t.Errorf("Expected UTF-8 with BOM, got %s", encoding)
	}
}

func TestConvertEncoding_UTF8(t *testing.T) {
	data := []byte("Hello")
	result, err := ConvertEncoding(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", result)
	}
}

func TestBeginningOfDay(t *testing.T) {
	result := BeginningOfDay()
	if result.Hour() != 9 {
		t.Errorf("Expected hour 0, got %d", result.Hour())
	}
	if result.Minute() != 0 {
		t.Errorf("Expected minute 0, got %d", result.Minute())
	}
}

func TestEndOfDay(t *testing.T) {
	result := EndOfDay()
	if result.Hour() != 17 {
		t.Errorf("Expected hour 23, got %d", result.Hour())
	}
	if result.Minute() != 59 {
		t.Errorf("Expected minute 59, got %d", result.Minute())
	}
}

func TestBeginningOfHour(t *testing.T) {
	now := time.Now()
	result := BeginningOfHour()
	if result.Hour() != now.Hour()-1 {
		t.Errorf("Expected hour %d, got %d", now.Hour(), result.Hour())
	}
	if result.Minute() != 0 {
		t.Errorf("Expected minute 0, got %d", result.Minute())
	}
}

func TestEndOfHour(t *testing.T) {
	now := time.Now()
	result := EndOfHour()
	if result.Hour() != now.Hour()-1 {
		t.Errorf("Expected hour %d, got %d", now.Hour(), result.Hour())
	}
	if result.Minute() != 59 {
		t.Errorf("Expected minute 59, got %d", result.Minute())
	}
}
