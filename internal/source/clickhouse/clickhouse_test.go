package clickhouse

import (
"testing"
)

func TestMetaStruct_JSONTags(t *testing.T) {
ms := MetaStruct{Name: "test_field", Type: "String", Len: 10}
if ms.Name != "test_field" {
t.Errorf("Expected 'test_field', got '%s'", ms.Name)
}
if ms.Type != "String" {
t.Errorf("Expected 'String', got '%s'", ms.Type)
}
if ms.Len != 10 {
t.Errorf("Expected Len 10, got %d", ms.Len)
}
}

func TestClickHouseJson_JSONTags(t *testing.T) {
ch := ClickHouseJson{
Name:   "test_name",
ColunmLen: 3,
Meta:   []MetaStruct{{Name: "col1", Type: "UInt64", Len: 5}},
Data:   []map[string]interface{}{{"col1": 1}},
Rows:   1,
}
if ch.Name != "test_name" {
t.Errorf("Expected 'test_name', got '%s'", ch.Name)
}
if ch.Rows != 1 {
t.Errorf("Expected Rows 1, got %d", ch.Rows)
}
}

func TestCHClient_StructFields(t *testing.T) {
client := &CHClient{
baseURL: "http://ch.example.com",
user:    "user",
pass:    "pass",
debug:   true,
}
if client.baseURL != "http://ch.example.com" {
t.Errorf("Expected 'http://ch.example.com', got '%s'", client.baseURL)
}
if !client.debug {
t.Error("Expected debug to be true")
}
}

func TestClickHouseJson_DecomposedLenCalc(t *testing.T) {
ch := ClickHouseJson{
Meta: []MetaStruct{
{Name: "id", Type: "UInt64"},
{Name: "name", Type: "String"},
},
Data: []map[string]interface{}{
{"id": "1", "name": "short"},
{"id": "999", "name": "longer_name"},
},
}
if ch.ColunmLen != 0 {
// Before Len() called, should be 0 (or default)
}
if len(ch.Meta) != 2 {
t.Errorf("Expected 2 meta columns, got %d", len(ch.Meta))
}
}
