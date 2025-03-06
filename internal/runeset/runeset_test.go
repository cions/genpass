// Copyright (c) 2024-2025 cions
// Licensed under the MIT License. See LICENSE for details.

package runeset_test

import (
	"fmt"
	"strings"
	"testing"
	"unicode"

	"github.com/cions/genpass/internal/runeset"
)

func assertEqual(t *testing.T, set runeset.RuneSet, want string, a ...any) {
	t.Helper()

	if got := set.String(); got != want {
		var prefix string
		if len(a) != 0 {
			prefix = fmt.Sprintf(a[0].(string), a[1:]...) + ": "
		}
		t.Errorf("%vexpected %v, but got %v", prefix, want, got)
	}
}

func TestRuneSet_Add(t *testing.T) {
	tests := []struct {
		char rune
		want string
	}{
		{'a', "a-ac-e"},
		{'b', "b-bc-e"},
		{'c', "c-e"},
		{'d', "c-e"},
		{'e', "c-e"},
		{'f', "c-ef-f"},
		{'g', "c-eg-g"},
	}

	for _, tt := range tests {
		var set runeset.RuneSet
		set.AddRange('c', 'e')
		set.Add(tt.char)
		assertEqual(t, set, tt.want, "Add(%q)", tt.char)
	}
}

func TestRuneSet_AddRange(t *testing.T) {
	t.Run("unit range", func(t *testing.T) {
		var set runeset.RuneSet
		set.AddRange('a', 'a')
		assertEqual(t, set, "a-a")
	})

	t.Run("lowercase", func(t *testing.T) {
		var set runeset.RuneSet
		set.AddRange('a', 'z')
		assertEqual(t, set, "a-z")
	})

	tests := []struct {
		lo, hi rune
		want   string
	}{
		{'a', 'a', "a-ac-eh-jk-kl-n"},
		{'a', 'c', "a-eh-jk-kl-n"},
		{'a', 'd', "a-eh-jk-kl-n"},
		{'a', 'e', "a-eh-jk-kl-n"},
		{'a', 'f', "a-fh-jk-kl-n"},
		{'a', 'h', "a-jk-kl-n"},
		{'a', 'k', "a-kl-n"},
		{'a', 'z', "a-z"},
		{'c', 'c', "c-eh-jk-kl-n"},
		{'c', 'd', "c-eh-jk-kl-n"},
		{'c', 'e', "c-eh-jk-kl-n"},
		{'c', 'f', "c-fh-jk-kl-n"},
		{'c', 'h', "c-jk-kl-n"},
		{'c', 'k', "c-kl-n"},
		{'c', 'l', "c-n"},
		{'c', 'z', "c-z"},
		{'f', 'f', "c-ef-fh-jk-kl-n"},
		{'f', 'g', "c-ef-gh-jk-kl-n"},
		{'f', 'h', "c-ef-jk-kl-n"},
		{'f', 'n', "c-ef-n"},
		{'f', 'z', "c-ef-z"},
		{'h', 'h', "c-eh-jk-kl-n"},
		{'h', 'j', "c-eh-jk-kl-n"},
		{'h', 'k', "c-eh-kl-n"},
		{'h', 'n', "c-eh-n"},
		{'i', 'j', "c-eh-jk-kl-n"},
		{'i', 'k', "c-eh-kl-n"},
		{'k', 'k', "c-eh-jk-kl-n"},
		{'k', 'l', "c-eh-jk-n"},
		{'k', 'm', "c-eh-jk-n"},
		{'k', 'n', "c-eh-jk-n"},
		{'k', 'z', "c-eh-jk-z"},
		{'x', 'x', "c-eh-jk-kl-nx-x"},
		{'x', 'z', "c-eh-jk-kl-nx-z"},
	}

	for _, tt := range tests {
		var set runeset.RuneSet
		set.AddRange('c', 'e')
		set.AddRange('h', 'j')
		set.AddRange('k', 'k')
		set.AddRange('l', 'n')
		set.AddRange(tt.lo, tt.hi)
		assertEqual(t, set, tt.want, "AddRange(%q, %q)", tt.lo, tt.hi)
	}
}

func TestRuneSet_AddRangeTable(t *testing.T) {
	table := &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x0041, Hi: 0x005A, Stride: 1},
			{Lo: 0x0061, Hi: 0x006A, Stride: 3},
		},
		R32: []unicode.Range32{
			{Lo: 0x10000, Hi: 0x10010, Stride: 1},
			{Lo: 0x10100, Hi: 0x10110, Stride: 16},
		},
		LatinOffset: 2,
	}

	var set runeset.RuneSet
	set.AddRangeTable(table)
	assertEqual(t, set, "A-Za-ad-dg-gj-j\U00010000-\U00010010\U00010100-\U00010100\U00010110-\U00010110")
}

func TestRuneSet_MergeAdjacents(t *testing.T) {
	var set runeset.RuneSet
	set.AddRange('a', 'c')
	set.AddRange('g', 'i')
	set.AddRange('j', 'j')
	set.AddRange('k', 'l')
	set.AddRange('s', 't')
	set.AddRange('u', 'v')
	set.AddRange('x', 'z')
	set.MergeAdjacents()
	assertEqual(t, set, "a-cg-ls-vx-z")
}

func TestRuneSet_Picker(t *testing.T) {
	expected := "abceghijklxyz"

	var set runeset.RuneSet
	set.AddRange('a', 'c')
	set.AddRange('e', 'e')
	set.AddRange('g', 'i')
	set.AddRange('j', 'j')
	set.AddRange('k', 'l')
	set.AddRange('x', 'z')
	set.MergeAdjacents()
	picker := set.Picker()

	if got := picker.Size(); got != int64(len(expected)) {
		t.Errorf("expected %v, but got %v", len(expected), got)
	}

	chars := make([]byte, picker.Size())
	for i := range picker.Size() {
		chars[i] = byte(picker.Get(i))
	}
	if got := string(chars); got != expected {
		t.Errorf("expected %v, but got %v", expected, got)
	}

	if r := picker.Random(); !strings.ContainsRune(expected, r) {
		t.Errorf("Random() returned a non-member rune %q", r)
	}
}
