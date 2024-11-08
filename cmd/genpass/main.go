// Copyright (c) 2024 cions
// Licensed under the MIT License. See LICENSE for details.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/cions/genpass/wordlists"
	"github.com/cions/go-colorterm"
	"github.com/cions/go-options"
	"github.com/cions/go-runeset"
)

var NAME = "genpass"
var VERSION = "(devel)"
var USAGE = `Usage: $NAME [-e] [-c N] [-w WORDLIST | -p | -x | -u] [-b BITS | -l N]

Generates secure random passphrases/password/hex/base64 strings.

Options:
  -e, --show-bits       Show the password strength
  -c, --count=N         Generate N strings
  -b, --bits=BITS       Generate strings with at least BITS-bit strength
                        (default: 80-bit for passphrase/password,
                                  128-bit for hex/base64)
  -l, --length=N        Generate N-words/characters strings
  -w, --wordlist={eff-large|eff-short1|eff-short2|bip39|slip39|FILE}
                        Generate passphrases using the specified wordlist
                        (default: eff-large)
  -p, --password        Generate passwords using ASCII graphical characters
      --password-with=CSET
                        Generate passwords using characters specified by CSET
  -x, --hex             Generate hexadecimal strings
  -u, --base64          Generate base64url strings
  -h, --help            Show this help message and exit
      --version         Show version information and exit

Syntax of CSET:
        c               Character c
        \-              Literal -
        \\              Literal \
        \xXX            Unicode character U+00XX
        \uXXXX          Unicode character U+XXXX
        \UXXXXXXXX      Unicode character U+XXXXXXXX
        c1-c2           Characters between c1 and c2 inclusive
        \d              ASCII digits
        \l              ASCII lowercase letters
        \L              ASCII uppercase letters
        \w              ASCII alphanumerics
        \&              ASCII punctuations
        \g              AScII graphical characters
        \pN             Unicode character class (one-letter General Category)
        \p{NAME}        Unicode character class (General Category or Scripts)
`

type Variant int

const (
	Passphrase Variant = iota
	Password
	Hexadecimal
	Base64
)

type Command struct {
	ShowBits bool
	Count    uint
	Variant  Variant
	Bits     uint
	Length   uint
	Wordlist string
	Picker   *runeset.Picker
}

func (c *Command) Kind(name string) options.Kind {
	switch name {
	case "-e", "--show-bits":
		return options.Boolean
	case "-c", "--count":
		return options.Required
	case "-b", "--bits":
		return options.Required
	case "-l", "--length":
		return options.Required
	case "-w", "--wordlist":
		return options.Required
	case "-p", "--password":
		return options.Boolean
	case "--password-with":
		return options.Required
	case "-x", "--hex":
		return options.Boolean
	case "-u", "--base64":
		return options.Boolean
	case "-h", "--help":
		return options.Boolean
	case "--version":
		return options.Boolean
	default:
		return options.Unknown
	}
}

func (c *Command) Option(name string, value string, hasValue bool) error {
	switch name {
	case "-e", "--show-bits":
		c.ShowBits = true
	case "-c", "--count":
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return err
		} else if n == 0 {
			return strconv.ErrRange
		}
		c.Count = uint(n)
	case "-b", "--bits":
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return err
		} else if n == 0 {
			return strconv.ErrRange
		}
		c.Bits = uint(n)
	case "-l", "--length":
		n, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return err
		} else if n == 0 {
			return strconv.ErrRange
		}
		c.Length = uint(n)
	case "-w", "--wordlist":
		c.Variant = Passphrase
		c.Wordlist = value
	case "-p", "--password":
		c.Variant = Password
		set, err := runeset.Parse(`\g`)
		if err != nil {
			return err
		}
		picker := set.Picker()
		if picker.Size() < 2 {
			return errors.New("must contain at least 2 characters")
		}
		c.Picker = picker
	case "--password-with":
		c.Variant = Password
		set, err := runeset.Parse(value)
		if err != nil {
			return err
		}
		picker := set.Picker()
		if picker.Size() < 2 {
			return errors.New("must contain at least 2 characters")
		}
		c.Picker = picker
	case "-x", "--hex":
		c.Variant = Hexadecimal
	case "-u", "--base64":
		c.Variant = Base64
	case "-h", "--help":
		return options.ErrHelp
	case "--version":
		return options.ErrVersion
	default:
		return options.ErrUnknown
	}
	return nil
}

func (c *Command) getWordlist() ([]string, error) {
	switch c.Wordlist {
	case "eff-large":
		return wordlists.EFFLarge, nil
	case "eff-short1":
		return wordlists.EFFShort1, nil
	case "eff-short2":
		return wordlists.EFFShort2, nil
	case "bip39":
		return wordlists.BIP39, nil
	case "slip39":
		return wordlists.SLIP39, nil
	}

	var r io.Reader = os.Stdin
	if c.Wordlist != "-" {
		f, err := os.Open(c.Wordlist)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	}

	var wordlist []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(wordlist) < 2 {
		return nil, errors.New("wordlist must contain at least 2 words")
	}
	return wordlist, nil
}

func (c *Command) getNumOfElems(bitsPerElem float64, defaultBits uint) uint {
	if c.Length != 0 {
		return c.Length
	} else if c.Bits != 0 {
		return uint(math.Ceil(float64(c.Bits) / bitsPerElem))
	} else {
		return uint(math.Ceil(float64(defaultBits) / bitsPerElem))
	}
}

func (c *Command) getGenerator() (Generator, float64, error) {
	switch c.Variant {
	case Passphrase:
		wordlist, err := c.getWordlist()
		if err != nil {
			return nil, 0, err
		}
		bitsPerElem := math.Log2(float64(len(wordlist)))
		nwords := c.getNumOfElems(bitsPerElem, 80)
		return newPassphraseGenerator(wordlist, nwords), bitsPerElem * float64(nwords), nil
	case Password:
		if c.Picker == nil {
			panic("genpass: c.Picker is nil")
		}
		bitsPerElem := math.Log2(float64(c.Picker.Size()))
		nchars := c.getNumOfElems(bitsPerElem, 80)
		return newPasswordGenerator(c.Picker, nchars), bitsPerElem * float64(nchars), nil
	case Hexadecimal:
		bitsPerElem := float64(4)
		nchars := c.getNumOfElems(bitsPerElem, 128)
		return newHexGenerator(nchars), bitsPerElem * float64(nchars), nil
	case Base64:
		bitsPerElem := float64(6)
		nchars := c.getNumOfElems(bitsPerElem, 128)
		return newBase64Generator(nchars), bitsPerElem * float64(nchars), nil
	default:
		panic("genpass: invalid Variant")
	}
}

func run(args []string) error {
	c := &Command{
		Count:    1,
		Variant:  Passphrase,
		Wordlist: "eff-large",
	}

	_, err := options.Parse(c, args)
	if errors.Is(err, options.ErrHelp) {
		usage := strings.ReplaceAll(USAGE, "$NAME", NAME)
		fmt.Print(usage)
		return nil
	} else if errors.Is(err, options.ErrVersion) {
		version := VERSION
		if bi, ok := debug.ReadBuildInfo(); ok {
			version = bi.Main.Version
		}
		fmt.Printf("%v %v\n", NAME, version)
		return nil
	} else if err != nil {
		return err
	}

	generator, bits, err := c.getGenerator()
	if err != nil {
		return err
	}

	for range c.Count {
		fmt.Print(generator())
		if c.ShowBits {
			fmt.Printf("\t\t%v(%.2f bits)%v", colorterm.Fg256Color(245), bits, colorterm.Reset)
		}
		fmt.Println()
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v: error: %v\n", NAME, err)
		os.Exit(1)
	}
}
