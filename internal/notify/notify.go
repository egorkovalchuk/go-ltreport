//go:build !notify_custom
// +build !notify_custom

package notify

func SendMessage(title, body string) error {
	return nil
}

func SendMessageWithAttach(title, body string, attachment string) error {
	return nil
}
