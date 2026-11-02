package hpsm

import (
"testing"
)

func TestContent_Struct(t *testing.T) {
content := Content{
Content: []struct {
AvrName          string `json:"avrName"`
BriefDescription string `json:"briefDescription"`
Description      string `json:"description"`
Number           string `json:"number"`
}{
{AvrName: "CPU", BriefDescription: "CPU Usage", Description: "CPU utilization", Number: "INC-001"},
},
}
if len(content.Content) != 1 {
t.Fatalf("Expected 1 item, got %d", len(content.Content))
}
if content.Content[0].Number != "INC-001" {
t.Errorf("Expected 'INC-001', got '%s'", content.Content[0].Number)
}
}

func TestFSMClient_Struct(t *testing.T) {
client := &FSMClient{
baseURL:  "https://fsm.example.com/api",
token:    "token123",
user:     "admin",
password: "secret",
period:   "PeriodStart=01-01-2024T00:00:00&PeriodEnd=02-01-2024T00:00:00",
}
if client.baseURL != "https://fsm.example.com/api" {
t.Errorf("Expected 'https://fsm.example.com/api', got '%s'", client.baseURL)
}
}

func TestFSMClient_EmptyContent(t *testing.T) {
content := Content{}
if len(content.Content) != 0 {
t.Errorf("Expected empty content, got %d items", len(content.Content))
}
}
