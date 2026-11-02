package allure

import (
"os"
"path/filepath"
"testing"
)

func TestHasTrailingSeparator_Empty(t *testing.T) {
if HasTrailingSeparator("") {
t.Error("Expected false for empty string")
}
}

func TestHasTrailingSeparator_WithSeparator(t *testing.T) {
if !HasTrailingSeparator("/home/") {
t.Error("Expected true for path with separator")
}
}

func TestHasTrailingSeparator_NoSeparator(t *testing.T) {
if HasTrailingSeparator("/home/user") {
t.Error("Expected false for path without trailing separator")
}
}

func TestEnsureTrailingSeparator_AddsSeparator(t *testing.T) {
result := EnsureTrailingSeparator("/home/user")
expected := "/home/user" + string(filepath.Separator)
if result != expected {
t.Errorf("Expected '%s', got '%s'", expected, result)
}
}

func TestEnsureTrailingSeparator_NoAddNeeded(t *testing.T) {
result := EnsureTrailingSeparator("/home/user/")
expected := "/home/user/"
if result != expected {
t.Errorf("Expected '%s', got '%s'", expected, result)
}
}

func TestEnsureTrailingSeparator_Empty(t *testing.T) {
result := EnsureTrailingSeparator("")
if result != string(filepath.Separator) {
t.Errorf("Expected separator for empty string, got '%s'", result)
}
}

func TestCreateOutputDir(t *testing.T) {
a := &Allure{
path: "/tmp/allure_test_dir_" + string(rune(os.Getpid())),
}
a.createOutputDir()
defer os.RemoveAll(a.path)
info, err := os.Stat(a.path)
if err != nil {
t.Errorf("Directory should exist: %v", err)
}
if !info.IsDir() {
t.Error("Path should be a directory")
}
}

func TestExistsDir_NonExistent(t *testing.T) {
a := &Allure{path: "/tmp/nonexistent_allure_test_dir_xyz"}
exists, err := a.existsDir(a.path)
if err != nil {
t.Errorf("Expected no error for non-existent dir check, got %v", err)
}
if exists {
t.Error("Expected false for non-existent directory")
}
}
