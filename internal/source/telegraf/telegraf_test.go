package telegraf

import (
"testing"
)

func TestTelegrafClient_Struct(t *testing.T) {
client := &TelegrafClient{
baseURL: "http://telegraf.example.com",
auth:    "Bearer token",
debug:   true,
}
if client.baseURL != "http://telegraf.example.com" {
t.Errorf("Expected 'http://telegraf.example.com', got '%s'", client.baseURL)
}
}
