// Copyright (c) 2024-2025 cions
// Licensed under the MIT License. See LICENSE for details.

package runeset_test

import (
	"testing"
	"unicode"

	"github.com/cions/genpass/internal/runeset"
)

func uniCharClass(table *unicode.RangeTable) string {
	var set runeset.RuneSet
	set.AddRangeTable(table)
	set.MergeAdjacents()
	return set.String()
}

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{``, ""},
		{`a`, "a-a"},
		{`\-`, "---"},
		{`\\`, "\\-\\"},
		{`\0`, "\u0000-\u0000"},
		{`\a`, "\u0007-\u0007"},
		{`\b`, "\u0008-\u0008"},
		{`\t`, "\u0009-\u0009"},
		{`\n`, "\u000A-\u000A"},
		{`\v`, "\u000B-\u000B"},
		{`\f`, "\u000C-\u000C"},
		{`\r`, "\u000D-\u000D"},
		{`\e`, "\u001B-\u001B"},
		{`\xFF`, "\u00FF-\u00FF"},
		{`\u3042`, "あ-あ"},
		{`\U0001F200`, "🈀-🈀"},
		{`ABCabc012`, "0-2A-Ca-c"},
		{`A-Ca-c0-2`, "0-2A-Ca-c"},
		{`a-zA-Z0-A`, "0-Za-z"},
		{`ぁ-ゖ`, "ぁ-ゖ"},
		{`ぁ-\u3096`, "ぁ-ゖ"},
		{`\u3041-ゖ`, "ぁ-ゖ"},
		{`\u3041-\u3096`, "ぁ-ゖ"},
		{`\U00020000-\U0002A6DF`, "\U00020000-\U0002A6DF"},
		{`\d`, "0-9"},
		{`\l`, "a-z"},
		{`\L`, "A-Z"},
		{`\w`, "0-9A-Za-z"},
		{`\s`, "!-/:-@[-`{-~"},
		{`\g`, "!-~"},
		{`\pL`, uniCharClass(unicode.L)},
		{`\p{Hiragana}`, uniCharClass(unicode.Hiragana)},
		{`\w\s\g\p{Lo}`, "!-~" + uniCharClass(unicode.Lo)},
		{`-a`, "---a-a"},
		{`a-`, "---a-a"},
		{`a\-z`, "---a-az-z"},
		{`a\\-z`, "\\-z"},
		{`!--/`, "!--/-/"},
		{`\w-_`, "---0-9A-Z_-_a-z"},
		{`--\d-\L--`, "---0-9A-Z"},
	}
	for _, tt := range tests {
		s, err := runeset.Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error: %v", tt.input, err)
		} else if got := s.String(); got != tt.want {
			t.Errorf("Parse(%q): expected %v, but got %v", tt.input, tt.want, got)
		}
	}
}

func TestParse_errors(t *testing.T) {
	tests := []string{
		`\`,
		`\?`,
		`\x`,
		`\x0`,
		`\xXX`,
		`\u`,
		`\u00`,
		`\uXXXX`,
		`\U`,
		`\U0000`,
		`\UXXXXXXXX`,
		`\p`,
		`\pX`,
		`\p{`,
		`\p{}`,
		`\p{Greek`,
		`\p{INVALID}`,
		`z-a`,
	}

	for _, tt := range tests {
		if _, err := runeset.Parse(tt); err == nil {
			t.Errorf("Parse(%q): expected a non-nil error", tt)
		}
	}
}
