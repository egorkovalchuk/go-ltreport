package reportdata

import (
"testing"
)

func TestCheckedWaitGroup_AddAndDone(t *testing.T) {
wg := &CheckedWaitGroup{}
wg.Add(3)
if wg.Count() != 3 {
t.Errorf("Expected count 3, got %d", wg.Count())
}
wg.Done()
wg.Done()
wg.Done()
if wg.Count() != 0 {
t.Errorf("Expected count 0, got %d", wg.Count())
}
}

func TestCheckedWaitGroup_PanicOnNegativeDelta(t *testing.T) {
wg := &CheckedWaitGroup{}
defer func() {
if r := recover(); r == nil {
t.Error("Expected panic on negative delta")
}
}()
wg.Add(-1)
}

func TestCheckedWaitGroup_PanicOnExtraDone(t *testing.T) {
wg := &CheckedWaitGroup{}
wg.Add(1)
defer func() {
if r := recover(); r == nil {
t.Error("Expected panic on extra Done")
}
}()
wg.Done()
wg.Done()
}

func TestCheckedWaitGroup_ExpectAtLeast(t *testing.T) {
wg := &CheckedWaitGroup{}
wg.Add(5)
if !wg.ExpectAtLeast(3) {
t.Error("Expected ExpectAtLeast(3) to be true")
}
if wg.ExpectAtLeast(10) {
t.Error("Expected ExpectAtLeast(10) to be false")
}
}
