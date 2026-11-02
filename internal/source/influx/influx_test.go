package influx

import (
	"testing"
)

func TestScenarioDinamic_Setters(t *testing.T) {
	s := &ScenarioDinamic{}
	s.SetApplication("app1")
	if s.NameTest != "app1" {
		t.Errorf("Expected NameTest 'app1', got '%s'", s.NameTest)
	}
	s.SetThread("thread1")
	if s.NameThread != "thread1" {
		t.Errorf("Expected NameThread 'thread1', got '%s'", s.NameThread)
	}
	field := []YField{{Name: "field1", Value: 10.0}}
	s.SeField(field)
	if len(s.Field) != 1 {
		t.Errorf("Expected 1 field, got %d", len(s.Field))
	}
}

func TestAddMap(t *testing.T) {
	m := make(map[string]map[string]KeyField)
	m = AddMap(m, "test1", "error1", KeyField{Value: 0.95, Description: "Error rate"})
	val, ok := m["test1"]["error1"]
	if !ok {
		t.Fatal("Expected key 'test1/error1' to exist")
	}
	if val.Value != 0.95 {
		t.Errorf("Expected Value 0.95, got %f", val.Value)
	}
	if val.Description != "Error rate" {
		t.Errorf("Expected Description 'Error rate', got '%s'", val.Description)
	}
}

func TestAddMapS(t *testing.T) {
	m := make(map[string]map[string]ScenarioDinamic)
	sd := ScenarioDinamic{NameTest: "test", NameThread: "thread1"}
	m = AddMapS(m, "test", "thread1", sd)
	val, ok := m["test"]["thread1"]
	if !ok {
		t.Fatal("Expected key 'test/thread1' to exist")
	}
	if val.NameTest != "test" {
		t.Errorf("Expected NameTest 'test', got '%s'", val.NameTest)
	}
}

func TestLTTestDinamic_NameTest(t *testing.T) {
	lt := LTTestDinamic{NameTest: "my-test"}
	if lt.NameTest != "my-test" {
		t.Errorf("Expected 'my-test', got '%s'", lt.NameTest)
	}
}
