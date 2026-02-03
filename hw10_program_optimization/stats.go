package hw10programoptimization

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	domainSuffix := "." + strings.ToLower(domain)
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	emailKey := []byte(`"Email":"`)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Быстрый поиск начала email
		start := bytes.Index(line, emailKey)
		if start == -1 {
			continue
		}

		valueStart := start + len(emailKey)
		if valueStart >= len(line) {
			continue
		}

		// Ищем закрывающую кавычку
		end := valueStart
		for end < len(line) && line[end] != '"' {
			end++
		}
		if end <= valueStart {
			continue
		}

		email := line[valueStart:end]
		at := bytes.LastIndexByte(email, '@')
		if at == -1 {
			continue
		}

		domainPart := email[at+1:]
		if len(domainPart) == 0 {
			continue
		}

		domainStr := strings.ToLower(string(domainPart))

		// Проверка суффикса
		if len(domainStr) <= len(domainSuffix) {
			continue
		}
		if domainStr[len(domainStr)-len(domainSuffix):] != domainSuffix {
			continue
		}

		result[domainStr]++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return result, nil
}
