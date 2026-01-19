package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

// Unpack выполняет примитивную распаковку строки вида "a4bc2d5e".
func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	var b strings.Builder
	var prevRune rune
	var escaping bool
	runes := []rune(input)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		switch {
		case escaping:
			if !(unicode.IsDigit(r) || r == '\\') {
				return "", ErrInvalidString
			}
			prevRune = r
			b.WriteRune(r)
			escaping = false

		case r == '\\':
			escaping = true

		case unicode.IsDigit(r):
			if prevRune == 0 {
				return "", ErrInvalidString
			}
			if err := repeatRune(&b, prevRune, r); err != nil {
				return "", err
			}
			prevRune = 0

		default:
			b.WriteRune(r)
			prevRune = r
		}
	}

	if escaping {
		return "", ErrInvalidString
	}

	return b.String(), nil
}

// repeatRune обрабатывает повторение символа в зависимости от цифры.
func repeatRune(b *strings.Builder, r rune, digit rune) error {
	count, err := strconv.Atoi(string(digit))
	if err != nil {
		return err
	}
	if count == 0 {
		// удалить последний символ
		output := []rune(b.String())
		if len(output) > 0 {
			output = output[:len(output)-1]
			b.Reset()
			b.WriteString(string(output))
		}
		return nil
	}
	b.WriteString(strings.Repeat(string(r), count-1))
	return nil
}
