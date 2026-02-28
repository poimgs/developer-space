package telegram

import "strings"

// specialChars are the characters that must be escaped in Telegram MarkdownV2.
// See: https://core.telegram.org/bots/api#markdownv2-style
var specialChars = []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

// EscapeMarkdownV2 escapes all MarkdownV2 special characters in the given text.
func EscapeMarkdownV2(text string) string {
	result := text
	// Backslash must be escaped first to avoid double-escaping.
	result = strings.ReplaceAll(result, `\`, `\\`)
	for _, ch := range specialChars {
		result = strings.ReplaceAll(result, ch, `\`+ch)
	}
	return result
}
