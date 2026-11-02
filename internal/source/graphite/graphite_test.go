package graphite

import (
"testing"
)

func TestDatapoint_Struct(t *testing.T) {
dp := Datapoint{Value: 42.5, Timestamp: 1700000000}
if dp.Value != 42.5 {
t.Errorf("Expected Value 42.5, got %f", dp.Value)
}
if dp.Timestamp != 1700000000 {
t.Errorf("Expected Timestamp 1700000000, got %d", dp.Timestamp)
}
}

func TestMetricResponse_Struct(t *testing.T) {
resp := MetricResponse{
Target: "cpu.usage",
Datapoints: []Datapoint{
{Value: 10.0, Timestamp: 1000},
{Value: 20.0, Timestamp: 2000},
},
}
if resp.Target != "cpu.usage" {
t.Errorf("Expected 'cpu.usage', got '%s'", resp.Target)
}
if len(resp.Datapoints) != 2 {
t.Errorf("Expected 2 datapoints, got %d", len(resp.Datapoints))
}
}

func TestGraphiteClient_Struct(t *testing.T) {
client := &GraphiteClient{
baseURL: "http://graphite.example.com",
auth:    "Bearer token",
debug:   true,
}
if client.baseURL != "http://graphite.example.com" {
t.Errorf("Expected 'http://graphite.example.com', got '%s'", client.baseURL)
}
}
