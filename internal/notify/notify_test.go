package notify

import (
	"testing"
)

func TestSendMessage_NoError(t *testing.T) {
	err := SendMessage("Test Title", "Test Body")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestSendMessageWithAttach_NoError(t *testing.T) {
	err := SendMessageWithAttach("Test Title", "Test Body", "/tmp/test.pdf")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestEscapeMarkdownV2_EmptyString(t *testing.T) {
	result := EscapeMarkdownV2("")
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestEscapeMarkdownV2_NoSpecialChars(t *testing.T) {
	input := "Hello World"
	result := EscapeMarkdownV2(input)
	expected := "Hello World"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEscapeMarkdownV2_SingleChar(t *testing.T) {
	input := "_bold_"
	result := EscapeMarkdownV2(input)
	expected := "\\_bold\\_"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEscapeMarkdownV2_MultipleChars(t *testing.T) {
	input := "[link](url) *bold* _italic_"
	result := EscapeMarkdownV2(input)
	expected := "\\[link\\]\\(url\\) \\*bold\\* \\_italic\\_"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEscapeMarkdownV2_NoChangePlain(t *testing.T) {
	input := "12345"
	result := EscapeMarkdownV2(input)
	// цифры должны остаться без изменений
	if result != "12345" {
		t.Errorf("Expected '12345', got '%s'", result)
	}
}
