//go:build !notify_custom
// +build !notify_custom

package notify

import "strings"

func SendMessage(title, body string) error {
	return nil
}

func SendMessageWithAttach(title, body string, attachment string) error {
	return nil
}

func EscapeMarkdownV2(text string) string {
	if text == "" {
		return ""
	}

	escapeChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#",
		"+", "-", "=", "|", "{", "}", ".", "!",
	}

	result := text
	for _, char := range escapeChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}

	return result
}
