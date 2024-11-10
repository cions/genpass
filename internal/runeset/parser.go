// Copyright (c) 2024 cions
// Licensed under the MIT License. See LICENSE for details.

package runeset

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func decodeCharClass(set *RuneSet, s string) (int, error) {
	if len(s) < 2 || s[0] != '\\' {
		return 0, nil
	}
	switch s[1] {
	case 'd':
		set.AddRange('0', '9')
		return 2, nil
	case 'l':
		set.AddRange('a', 'z')
		return 2, nil
	case 'L':
		set.AddRange('A', 'Z')
		return 2, nil
	case 'w':
		set.AddRange('0', '9')
		set.AddRange('A', 'Z')
		set.AddRange('a', 'z')
		return 2, nil
	case 's':
		set.AddRange('!', '/')
		set.AddRange(':', '@')
		set.AddRange('[', '`')
		set.AddRange('{', '~')
		return 2, nil
	case 'g':
		set.AddRange('!', '~')
		return 2, nil
	case 'p':
		if len(s) < 3 {
			return 0, fmt.Errorf("truncated escape sequence: %s", s)
		}
		if s[2] != '{' {
			if table, ok := unicode.Categories[string(s[2])]; ok {
				set.AddRangeTable(table)
			} else {
				return 0, fmt.Errorf("invalid character class name: %s", s[:3])
			}
			return 3, nil
		}
		end := strings.IndexByte(s, '}')
		if end < 0 {
			return 0, fmt.Errorf("unterminated escape sequence: %s", s)
		}
		name := s[3:end]
		if table, ok := unicode.Categories[name]; ok {
			set.AddRangeTable(table)
		} else if table, ok := unicode.Scripts[name]; ok {
			set.AddRangeTable(table)
		} else {
			return 0, fmt.Errorf("invalid character class name: %s", s[:end+1])
		}
		return end + 1, nil
	default:
		return 0, nil
	}
}

func decodeChar(s string) (rune, int, error) {
	if len(s) == 0 {
		return 0, 0, io.EOF
	}
	if s[0] != '\\' {
		r, size := utf8.DecodeRuneInString(s)
		return r, size, nil
	}
	if len(s) == 1 {
		return 0, 0, fmt.Errorf("truncated escape sequence: %s", s)
	}
	switch s[1] {
	case '-', '\\':
		return rune(s[1]), 2, nil
	case '0':
		return '\x00', 2, nil
	case 'a':
		return '\x07', 2, nil
	case 'b':
		return '\x08', 2, nil
	case 't':
		return '\x09', 2, nil
	case 'n':
		return '\x0A', 2, nil
	case 'v':
		return '\x0B', 2, nil
	case 'f':
		return '\x0C', 2, nil
	case 'r':
		return '\x0D', 2, nil
	case 'e':
		return '\x1B', 2, nil
	case 'x':
		if len(s) < 4 {
			return 0, 0, fmt.Errorf("truncated escape sequence: %s", s)
		}
		n, err := strconv.ParseUint(s[2:4], 16, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid escape sequence: %s", s[:4])
		}
		return rune(n), 4, nil
	case 'u':
		if len(s) < 6 {
			return 0, 0, fmt.Errorf("truncated escape sequence: %s", s)
		}
		n, err := strconv.ParseUint(s[2:6], 16, 32)
		if err != nil || !utf8.ValidRune(rune(n)) {
			return 0, 0, fmt.Errorf("invalid escape sequence: %s", s[:6])
		}
		return rune(n), 6, nil
	case 'U':
		if len(s) < 10 {
			return 0, 0, fmt.Errorf("truncated escape sequence: %s", s)
		}
		n, err := strconv.ParseUint(s[2:10], 16, 32)
		if err != nil || !utf8.ValidRune(rune(n)) {
			return 0, 0, fmt.Errorf("invalid escape sequence: %s", s[:10])
		}
		return rune(n), 10, nil
	default:
		return 0, 0, fmt.Errorf("invalid escape sequence: %s", s[:2])
	}
}

func Parse(s string) (RuneSet, error) {
	var set RuneSet

	for len(s) != 0 {
		if size, err := decodeCharClass(&set, s); err != nil {
			return RuneSet{}, err
		} else if size != 0 {
			s = s[size:]
			continue
		}
		r, size, err := decodeChar(s)
		if err != nil {
			return RuneSet{}, err
		}
		if len(s) > size && s[size] == '-' {
			end, endsize, err := decodeChar(s[size+1:])
			if err == nil {
				if r > end {
					return RuneSet{}, fmt.Errorf("bad character range: %s", s[:size+endsize+1])
				}
				set.AddRange(r, end)
				s = s[size+endsize+1:]
				continue
			}
		}
		set.Add(r)
		s = s[size:]
	}

	set.MergeAdjacents()
	return set, nil
}
