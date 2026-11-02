package grafana

import (
	"testing"
)

func TestGetMajorVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"10.2.3", 10},
		{"11.0.0", 11},
		{"9.5.1", 9},
		{"invalid", 0},
		{"", 0},
	}
	for _, tc := range tests {
		result := getMajorVersion(tc.input)
		if result != tc.expected {
			t.Errorf("getMajorVersion(%q) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

func TestAlertToString_Int(t *testing.T) {
	result, err := alertToString(42)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "42" {
		t.Errorf("Expected '42', got '%s'", result)
	}
}

func TestAlertToString_String(t *testing.T) {
	result, err := alertToString("abc123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "abc123" {
		t.Errorf("Expected 'abc123', got '%s'", result)
	}
}

func TestAlertToString_Nil(t *testing.T) {
	_, err := alertToString(nil)
	if err == nil {
		t.Error("Expected error for nil AlertID")
	}
}

func TestAlertToString_Float64(t *testing.T) {
	result, err := alertToString(3.14)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "3.14" {
		t.Errorf("Expected '3.14', got '%s'", result)
	}
}

func TestAlertToString_Int64(t *testing.T) {
	result, err := alertToString(int64(999))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result != "999" {
		t.Errorf("Expected '999', got '%s'", result)
	}
}

func TestConvertUnixMillisToTime(t *testing.T) {
	// 1000 milliseconds = 1 second
	result := convertUnixMillisToTime(1000)
	expectedSecond := int64(1)
	if result.Unix() != expectedSecond {
		t.Errorf("Expected unix %d, got %d", expectedSecond, result.Unix())
	}
}

func TestGrafanaClient_Struct(t *testing.T) {
	client := &GrafanaClient{
		baseURL: "http://grafana.example.com",
		auth:    "Bearer token",
		debug:   true,
	}
	if client.baseURL != "http://grafana.example.com" {
		t.Errorf("Expected 'http://grafana.example.com', got '%s'", client.baseURL)
	}
}
