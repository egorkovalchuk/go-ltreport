package allure

import (
"encoding/base64"
"encoding/json"
"testing"
"time"
)

func TestAllureResult_JSONMarshal(t *testing.T) {
now := time.Now()
ar := &AllureResult{
UUID:      "test-uuid-123",
Name:      "Test Case",
FullName:  "package.TestCase",
HistoryID: "abc123",
Status:    "passed",
Stage:     "finished",
Start:     now.UnixMilli(),
Stop:      now.Add(1 * time.Second).UnixMilli(),
}
ar.AddStep("Step 1", "passed", now, now.Add(500*time.Millisecond))
ar.AddLabel("layer", "unit")
ar.AddLink("Jira", "https://jira.example.com/TEST-1", "issue")

data, err := json.Marshal(ar)
if err != nil {
t.Fatalf("Unexpected marshal error: %v", err)
}

var unmarshaled AllureResult
err = json.Unmarshal(data, &unmarshaled)
if err != nil {
t.Fatalf("Unexpected unmarshal error: %v", err)
}
if unmarshaled.UUID != "test-uuid-123" {
t.Errorf("Expected UUID 'test-uuid-123', got '%s'", unmarshaled.UUID)
}
if unmarshaled.Status != "passed" {
t.Errorf("Expected status 'passed', got '%s'", unmarshaled.Status)
}
if len(unmarshaled.Steps) != 1 {
t.Errorf("Expected 1 step, got %d", len(unmarshaled.Steps))
}
}

func TestAllureAttachment_Base64(t *testing.T) {
content := []byte("Hello, World!")
encoded := base64.StdEncoding.EncodeToString(content)
var decoded []byte
decoded, err := base64.StdEncoding.DecodeString(encoded)
if err != nil {
t.Fatalf("Unexpected decode error: %v", err)
}
if string(decoded) != string(content) {
t.Errorf("Round-trip failed: expected '%s', got '%s'", content, decoded)
}
}

func TestAttachmentTypes_Constants(t *testing.T) {
tests := []struct {
name string
typ  AttachmentType
}{
{"Text", Text},
{"JSON", JSON},
{"XML", XML},
{"CSV", CSV},
{"HTML", HTML},
{"PNG", PNG},
{"JPEG", JPEG},
{"SVG", SVG},
{"Binary", Binary},
}
for _, tc := range tests {
if tc.typ == "" {
t.Errorf("AttachmentType %s should not be empty", tc.name)
}
}
}

func TestAllureParameter_JSONMarshal(t *testing.T) {
p := AllureParameter{Name: "browser", Value: "chrome"}
data, err := json.Marshal(p)
if err != nil {
t.Fatalf("Unexpected marshal error: %v", err)
}
var p2 AllureParameter
err = json.Unmarshal(data, &p2)
if err != nil {
t.Fatalf("Unexpected unmarshal error: %v", err)
}
if p2.Name != "browser" || p2.Value != "chrome" {
t.Errorf("Round-trip failed: name=%s value=%s", p2.Name, p2.Value)
}
}

func TestAllureLink_JSONMarshal(t *testing.T) {
l := AllureLink{Name: "Confluence", URL: "https://wiki.example.com/page", Type: "requirement"}
data, err := json.Marshal(l)
if err != nil {
t.Fatalf("Unexpected marshal error: %v", err)
}
var l2 AllureLink
err = json.Unmarshal(data, &l2)
if err != nil {
t.Fatalf("Unexpected unmarshal error: %v", err)
}
if l2.URL != "https://wiki.example.com/page" {
t.Errorf("Expected URL 'https://wiki.example.com/page', got '%s'", l2.URL)
}
}

func TestAllureStep_JSONMarshal(t *testing.T) {
now := time.Now()
s := AllureStep{
Name:   "Check Response",
Status: "passed",
Start:  now.UnixMilli(),
Stop:   now.Add(100 * time.Millisecond).UnixMilli(),
}
s2 := AllureStep{Name: "SubStep", Status: "passed", Start: now.UnixMilli(), Stop: now.Add(50 * time.Millisecond).UnixMilli()}
s.Steps = append(s.Steps, s2)
data, err := json.Marshal(s)
if err != nil {
t.Fatalf("Unexpected marshal error: %v", err)
}
var s3 AllureStep
err = json.Unmarshal(data, &s3)
if err != nil {
t.Fatalf("Unexpected unmarshal error: %v", err)
}
if len(s3.Steps) != 1 {
t.Errorf("Expected 1 nested step, got %d", len(s3.Steps))
}
}
