package reportdata

import (
	"testing"
)

// --- parseThreshold ---

func TestParseThreshold_StringGreaterThan(t *testing.T) {
	cfg := &Config{}
	th, com, err := cfg.parseThreshold(">100")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if th != 100 {
		t.Errorf("Expected threshold 100, got %f", th)
	}
	if com != ">" {
		t.Errorf("Expected comparison '>', got '%s'", com)
	}
}

func TestParseThreshold_StringLessThan(t *testing.T) {
	cfg := &Config{}
	th, com, err := cfg.parseThreshold("<50")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if th != 50 {
		t.Errorf("Expected threshold 50, got %f", th)
	}
	if com != "<" {
		t.Errorf("Expected comparison '<', got '%s'", com)
	}
}

func TestParseThreshold_Int(t *testing.T) {
	cfg := &Config{}
	th, com, err := cfg.parseThreshold(200)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if th != 200.0 {
		t.Errorf("Expected 200.0, got %f", th)
	}
	if com != ">" {
		t.Errorf("Expected '>' for int, got '%s'", com)
	}
}

func TestParseThreshold_Float64(t *testing.T) {
	cfg := &Config{}
	th, com, err := cfg.parseThreshold(3.14)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if th != 3.14 {
		t.Errorf("Expected 3.14, got %f", th)
	}
	if com != ">" {
		t.Errorf("Expected '>' for float, got '%s'", com)
	}
}

// --- LTErrors ---

func TestLTErrors_AddError(t *testing.T) {
	var errs LTErrors
	errs.AddError("test_metric", 95.0, "High latency", 98.5, "Grafana", "tag1")
	if len(errs) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errs))
	}
	if errs[0].Name != "Grafana: test_metric" {
		t.Errorf("Expected 'Grafana: test_metric', got '%s'", errs[0].Name)
	}
	if errs[0].Threshold != 95.0 {
		t.Errorf("Expected threshold 95.0, got %f", errs[0].Threshold)
	}
}

func TestLTErrors_MultipleErrors(t *testing.T) {
	var errs LTErrors
	errs.AddError("metric1", 10.0, "desc1", 15.0, "Influx", "tag1")
	errs.AddError("metric2", 20.0, "desc2", 25.0, "Prometheus", "tag2")
	if len(errs) != 2 {
		t.Fatalf("Expected 2 errors, got %d", len(errs))
	}
}
