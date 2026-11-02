package prometheus

import (
	"testing"
)

func TestPrometheusClient_Struct(t *testing.T) {
	client := &PrometheusClient{
		baseURL: "http://prometheus.example.com",
		debug:   true,
	}
	if client.baseURL != "http://prometheus.example.com" {
		t.Errorf("Expected 'http://prometheus.example.com', got '%s'", client.baseURL)
	}
}
