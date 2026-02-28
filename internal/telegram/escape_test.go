package telegram

import "testing"

func TestEscapeMarkdownV2_SpecialChars(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"underscore", "hello_world", `hello\_world`},
		{"asterisk", "hello*world", `hello\*world`},
		{"brackets", "[hello](world)", `\[hello\]\(world\)`},
		{"tilde", "hello~world", `hello\~world`},
		{"backtick", "hello`world", "hello\\`world"},
		{"greater than", "hello>world", `hello\>world`},
		{"hash", "hello#world", `hello\#world`},
		{"plus", "hello+world", `hello\+world`},
		{"minus", "hello-world", `hello\-world`},
		{"equals", "hello=world", `hello\=world`},
		{"pipe", "hello|world", `hello\|world`},
		{"braces", "hello{world}", `hello\{world\}`},
		{"dot", "hello.world", `hello\.world`},
		{"exclamation", "hello!world", `hello\!world`},
		{"backslash", `hello\world`, `hello\\world`},
		{"plain text", "Hello World", "Hello World"},
		{"empty string", "", ""},
		{"multiple specials", "a_b*c.d!", `a\_b\*c\.d\!`},
		{"time format", "14:00", `14:00`},
		{"date with dash", "2025-02-14", `2025\-02\-14`},
		{"rsvp count", "4 / 8", `4 / 8`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EscapeMarkdownV2(tt.input)
			if got != tt.want {
				t.Errorf("EscapeMarkdownV2(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeMarkdownV2_BackslashFirst(t *testing.T) {
	// Backslash must be escaped before other chars to avoid double-escaping.
	input := `\_already_escaped`
	got := EscapeMarkdownV2(input)
	want := `\\\_already\_escaped`
	if got != want {
		t.Errorf("EscapeMarkdownV2(%q) = %q, want %q", input, got, want)
	}
}
