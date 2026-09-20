// Package envfile loads environment variables from .env files.
package envfile

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

const DefaultFile = ".env"

var (
	escapePattern   = regexp.MustCompile(`\\.`)
	unescapePattern = regexp.MustCompile(`\\([^$])`)
	varPattern      = regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)
)

// Load reads .env files and sets env vars; errors silently discarded.
func Load(additionalFiles ...string) {
	files := make([]string, 0, 1+len(additionalFiles))
	files = append(files, DefaultFile)
	files = append(files, additionalFiles...)

	for _, f := range files {
		loadFile(f) //nolint:errcheck // intentional silent discard
	}
}

func loadFile(filename string) error {
	fi, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if fi.IsDir() {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	vars := make(map[string]string)
	if err := parseContent(data, vars); err != nil {
		return err
	}

	for k, v := range vars {
		os.Setenv(k, v)
	}

	return nil
}

func parseContent(src []byte, out map[string]string) error {
	src = bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))

	for {
		src = skipToStatement(src)
		if src == nil {
			return nil
		}

		key, remaining, err := extractKey(src)
		if err != nil {
			return err
		}

		value, rest, err := extractValue(remaining, out)
		if err != nil {
			return err
		}

		out[key] = value
		src = rest
	}
}

func skipToStatement(src []byte) []byte {
	i := bytes.IndexFunc(src, func(r rune) bool { return !unicode.IsSpace(r) })
	if i < 0 {
		return nil
	}
	src = src[i:]

	if len(src) > 0 && src[0] == '#' {
		j := bytes.IndexByte(src, '\n')
		if j < 0 {
			return nil
		}
		return skipToStatement(src[j+1:])
	}

	return src
}

func extractKey(src []byte) (key string, remaining []byte, err error) {
	src = bytes.TrimLeftFunc(src, isSpace)

	if bytes.HasPrefix(src, []byte("export")) && len(src) > 6 && isSpace(rune(src[6])) {
		src = src[6:]
		src = bytes.TrimLeftFunc(src, isSpace)
	}

	sep := -1
	for i, b := range src {
		if b == '=' || b == ':' {
			sep = i
			break
		}
		switch {
		case isSpace(rune(b)):
			continue
		case b == '_':
			continue
		case unicode.IsLetter(rune(b)) || unicode.IsNumber(rune(b)) || b == '.':
			continue
		default:
			return "", nil, fmt.Errorf("unexpected character %q in variable name near %q", b, string(src))
		}
	}

	if sep < 0 {
		if len(src) == 0 {
			return "", nil, errors.New("zero length string")
		}
		key = strings.TrimRightFunc(string(src), unicode.IsSpace)
		return key, nil, nil
	}

	key = string(src[:sep])
	remaining = src[sep+1:]
	key = strings.TrimRightFunc(key, unicode.IsSpace)
	remaining = bytes.TrimLeftFunc(remaining, isSpace)

	return
}

func extractValue(src []byte, vars map[string]string) (value string, rest []byte, err error) {
	if len(src) == 0 {
		return "", nil, nil
	}

	prefix, isQuoted := hasQuotePrefix(src)
	if !isQuoted {
		return extractUnquoted(src, vars)
	}

	if prefix == '\'' {
		return extractSingleQuoted(src)
	}

	return extractDoubleQuoted(src, vars)
}

func extractUnquoted(src []byte, vars map[string]string) (value string, rest []byte, err error) {
	i := bytes.IndexFunc(src, isLineEnd)
	var line []byte
	if i < 0 {
		line = src
		rest = nil
	} else {
		line = src[:i]
		rest = src[i:]
	}

	if len(line) == 0 {
		return "", rest, nil
	}

	runes := []rune(string(line))
	commentIdx := -1
	for j := len(runes) - 1; j >= 0; j-- {
		if runes[j] == '#' && j > 0 && isSpace(runes[j-1]) {
			commentIdx = j
			break
		}
	}
	if commentIdx >= 0 {
		line = []byte(string(runes[:commentIdx]))
	}

	trimmed := strings.TrimFunc(string(line), isSpace)

	trimmed = resolveVars(trimmed, vars)

	return trimmed, rest, nil
}

func extractSingleQuoted(src []byte) (value string, rest []byte, err error) {
	for i := 1; i < len(src); i++ {
		if src[i] == '\'' && src[i-1] != '\\' {
			value = string(src[1:i])
			rest = src[i+1:]
			return
		}
		if src[i] == '\n' {
			return "", nil, fmt.Errorf("unterminated quoted value %s", string(src))
		}
	}
	return "", nil, fmt.Errorf("unterminated quoted value %s", string(src))
}

func extractDoubleQuoted(src []byte, vars map[string]string) (value string, rest []byte, err error) {
	for i := 1; i < len(src); i++ {
		if src[i] == '"' && src[i-1] != '\\' {
			value = string(src[1:i])
			value = processEscapes(value)
			value = resolveVars(value, vars)
			rest = src[i+1:]
			return
		}
		if src[i] == '\n' {
			return "", nil, fmt.Errorf("unterminated quoted value %s", string(src))
		}
	}
	return "", nil, fmt.Errorf("unterminated quoted value %s", string(src))
}

func processEscapes(str string) string {
	str = escapePattern.ReplaceAllStringFunc(str, func(match string) string {
		if len(match) < 2 {
			return match
		}
		switch match[1] {
		case 'n':
			return "\n"
		case 'r':
			return "\r"
		default:
			return match
		}
	})

	str = unescapePattern.ReplaceAllString(str, "$1")
	return str
}

func resolveVars(v string, m map[string]string) string {
	return varPattern.ReplaceAllStringFunc(v, func(match string) string {
		subs := varPattern.FindStringSubmatch(match)

		if subs[1] == `\` || subs[3] == `(` {
			return match[1:]
		}

		if subs[4] != "" {
			if val, ok := m[subs[4]]; ok {
				return val
			}
			return ""
		}

		return match
	})
}

func hasQuotePrefix(src []byte) (prefix byte, isQuoted bool) {
	if len(src) == 0 {
		return 0, false
	}
	if src[0] == '\'' || src[0] == '"' {
		return src[0], true
	}
	return 0, false
}

func isSpace(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}

func isLineEnd(r rune) bool {
	return r == '\n' || r == '\r'
}
