package allure

import (
	"testing"
	"time"
)

func TestAllureResult_FilishedFailed(t *testing.T) {
	ar := &AllureResult{}
	ar.FilishedFailed()
	if ar.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", ar.Status)
	}
	if ar.Stage != "finished" {
		t.Errorf("Expected stage 'finished', got '%s'", ar.Stage)
	}
}

func TestAllureResult_FilishedPassed(t *testing.T) {
	ar := &AllureResult{}
	ar.FilishedPassed()
	if ar.Status != "passed" {
		t.Errorf("Expected status 'passed', got '%s'", ar.Status)
	}
}

func TestAllureResult_FilishedBroken(t *testing.T) {
	ar := &AllureResult{}
	ar.FilishedBroken()
	if ar.Status != "broken" {
		t.Errorf("Expected status 'broken', got '%s'", ar.Status)
	}
}

func TestAllureResult_AddStep(t *testing.T) {
	ar := &AllureResult{
		Start: time.Now().UnixMilli(),
		Stop:  time.Now().Add(1 * time.Second).UnixMilli(),
	}
	start := time.Now()
	end := time.Now().Add(500 * time.Millisecond)
	ar.AddStep("Test Step", "passed", start, end)
	if len(ar.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(ar.Steps))
	}
	if ar.Steps[0].Name != "Test Step" {
		t.Errorf("Expected name 'Test Step', got '%s'", ar.Steps[0].Name)
	}
	if ar.Steps[0].Status != "passed" {
		t.Errorf("Expected status 'passed', got '%s'", ar.Steps[0].Status)
	}
}

func TestAllureResult_AddLabel(t *testing.T) {
	ar := &AllureResult{}
	ar.AddLabel("layer", "integration")
	if len(ar.Labels) != 1 {
		t.Fatalf("Expected 1 label, got %d", len(ar.Labels))
	}
	if ar.Labels[0].Name != "layer" {
		t.Errorf("Expected name 'layer', got '%s'", ar.Labels[0].Name)
	}
	if ar.Labels[0].Value != "integration" {
		t.Errorf("Expected value 'integration', got '%s'", ar.Labels[0].Value)
	}
}

func TestAllureResult_AddLink(t *testing.T) {
	ar := &AllureResult{}
	ar.AddLink("Jira", "https://jira.example.com/TEST-123", "issue")
	if len(ar.Links) != 1 {
		t.Fatalf("Expected 1 link, got %d", len(ar.Links))
	}
	if ar.Links[0].Name != "Jira" {
		t.Errorf("Expected name 'Jira', got '%s'", ar.Links[0].Name)
	}
}

func TestAllureResult_AddStepWithParam(t *testing.T) {
	ar := &AllureResult{}
	start := time.Now()
	end := time.Now().Add(1 * time.Second)
	params := []AllureParameter{
		{Name: "key", Value: "value"},
		{Name: "key2", Value: "value2"},
	}
	ar.AddStepWithParam("Step with params", "passed", start, end, params)
	if len(ar.Steps) != 1 {
		t.Fatalf("Expected 1 step, got %d", len(ar.Steps))
	}
	if len(ar.Steps[0].Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(ar.Steps[0].Parameters))
	}
	if ar.Steps[0].Parameters[0].Name != "key" {
		t.Errorf("Expected parameter name 'key', got '%s'", ar.Steps[0].Parameters[0].Name)
	}
}

func TestAllureResult_GetValueLabel_Found(t *testing.T) {
	ar := &AllureResult{}
	ar.AddLabel("layer", "unit")
	ar.AddLabel("testType", "integration")
	val := ar.GetValueLabel("layer")
	if val != "unit" {
		t.Errorf("Expected 'unit', got '%s'", val)
	}
}

func TestAllureResult_GetValueLabel_NotFound(t *testing.T) {
	ar := &AllureResult{}
	val := ar.GetValueLabel("nonexistent")
	if val != "Unknown" {
		t.Errorf("Expected 'Unknown', got '%s'", val)
	}
}

func TestAllureResult_ArrayToParam(t *testing.T) {
	ar := &AllureResult{}
	params := ar.ArrayToParam("key1:value1;key2:value2")
	if len(params) != 2 {
		t.Fatalf("Expected 2 parameters, got %d", len(params))
	}
	if params[0].Name != "key1" {
		t.Errorf("Expected 'key1', got '%s'", params[0].Name)
	}
	if params[0].Value != "value1" {
		t.Errorf("Expected 'value1', got '%s'", params[0].Value)
	}
}

func TestAllureResult_ArrayToParam_Empty(t *testing.T) {
	ar := &AllureResult{}
	params := ar.ArrayToParam("")
	if len(params) != 0 {
		t.Errorf("Expected 0 parameters, got %d", len(params))
	}
}

func TestAllureResult_AddStepChild(t *testing.T) {
	ar := &AllureResult{}
	step := &AllureStep{}
	start := time.Now()
	end := time.Now().Add(1 * time.Second)
	params := []AllureParameter{{Name: "k", Value: "v"}}
	childSteps := []AllureStep{{Name: "child1"}}
	attachments := []AllureAttachment{{Name: "log.txt", Type: "text/plain"}}
	ar.AddStepChild("Parent", "passed", start, end, params, childSteps, attachments, step)
	if len(step.Steps) != 1 {
		t.Errorf("Expected 1 child step, got %d", len(step.Steps))
	}
}
