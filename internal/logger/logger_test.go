package logger

import (
"testing"
)

func TestNewLogWriter(t *testing.T) {
lw := NewLogWriter("/tmp/test_log.log", false)
if lw == nil {
t.Fatal("Expected non-nil LogWriter")
}
if lw.logFileName != "/tmp/test_log.log" {
t.Errorf("Expected logFileName '/tmp/test_log.log', got '%s'", lw.logFileName)
}
if lw.console {
t.Error("Expected console to be false by default")
}
if lw.debugm {
t.Error("Expected debugm to be false by default")
}
}

func TestLogWriter_SetConsoleOutput(t *testing.T) {
lw := NewLogWriter("/tmp/test_log.log", false)
lw.SetConsoleOutput(true)
if !lw.console {
t.Error("Expected console to be true")
}
lw.SetConsoleOutput(false)
if lw.console {
t.Error("Expected console to be false")
}
}

func TestLogWriter_ChangeDebugLevel(t *testing.T) {
lw := NewLogWriter("/tmp/test_log.log", false)
if lw.debugm {
t.Error("Expected debugm to be false")
}
lw.ChangeDebugLevel(true)
if !lw.debugm {
t.Error("Expected debugm to be true")
}
}
