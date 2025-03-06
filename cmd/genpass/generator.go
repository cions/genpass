// Copyright (c) 2024-2025 cions
// Licensed under the MIT License. See LICENSE for details.

package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/cions/genpass/internal/runeset"
)

type Generator func() string

func choice[S ~[]E, E any](slice S) E {
	n := big.NewInt(int64(len(slice)))
	i, err := rand.Int(rand.Reader, n)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	} else if !i.IsInt64() {
		panic("crypto/rand: out of range")
	}
	return slice[i.Int64()]
}

func newPassphraseGenerator(wordlist []string, nwords uint) Generator {
	if len(wordlist) == 0 {
		panic("newPassphraseGenerator: empty wordlist")
	}
	return func() string {
		words := make([]string, nwords)
		for i := range nwords {
			words[i] = choice(wordlist)
		}
		return strings.Join(words, " ")
	}
}

func newPasswordGenerator(picker *runeset.Picker, nchars uint) Generator {
	if picker.Size() == 0 {
		panic("newPasswordGenerator: empty runeset")
	}
	return func() string {
		chars := make([]string, nchars)
		for i := range nchars {
			chars[i] = string(picker.Random())
		}
		return strings.Join(chars, "")
	}
}

func newHexGenerator(nchars uint) Generator {
	if nchars == 0 {
		panic("newHexGenerator: nchars must not be zero")
	}
	return func() string {
		buf := make([]byte, (nchars-1)/2+1)
		if _, err := rand.Read(buf); err != nil {
			panic(fmt.Sprintf("crypto/rand: %v", err))
		}
		return hex.EncodeToString(buf)[:nchars]
	}
}

func newBase64Generator(nchars uint) Generator {
	if nchars == 0 {
		panic("newBase64Generator: nchars must not be zero")
	}
	return func() string {
		buf := make([]byte, 3*((nchars-1)/4+1))
		if _, err := rand.Read(buf); err != nil {
			panic(fmt.Sprintf("crypto/rand: %v", err))
		}
		return base64.URLEncoding.EncodeToString(buf)[:nchars]
	}
}
