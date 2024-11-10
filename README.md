# genpass

[![GitHub Releases](https://img.shields.io/github/v/release/cions/genpass?sort=semver)](https://github.com/cions/genpass/releases)
[![LICENSE](https://img.shields.io/github/license/cions/genpass)](https://github.com/cions/genpass/blob/master/LICENSE)
[![CI](https://github.com/cions/genpass/actions/workflows/ci.yml/badge.svg)](https://github.com/cions/genpass/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cions/genpass.svg)](https://pkg.go.dev/github.com/cions/genpass)
[![Go Report Card](https://goreportcard.com/badge/github.com/cions/genpass)](https://goreportcard.com/report/github.com/cions/genpass)

Generates secure random passphrases/password/hex/base64 strings.

## Usage

```
$ genpass --help
Usage: genpass [-e] [-c N] [-w WORDLIST | -p | -x | -u] [-b BITS | -l N]

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
```

## Installation

[Download from GitHub Releases](https://github.com/cions/genpass/releases)

### Build from source

```sh
$ go install github.com/cions/genpass/cmd/genpass@latest
```

## License

MIT
