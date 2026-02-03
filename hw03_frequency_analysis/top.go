package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var re = regexp.MustCompile(`[^\s]+`)

func Top10(text string) []string {
	if text == "" {
		return []string{}
	}

	// Разбиваем текст на слова с помощью регулярного выражения
	words := re.FindAllString(text, -1)

	// Считаем частоту слов с обработкой по дополнительному заданию
	freq := make(map[string]int)
	for _, word := range words {
		// Приводим к нижнему регистру
		processedWord := strings.ToLower(word)

		// Убираем знаки препинания с краёв слова
		processedWord = strings.TrimFunc(processedWord, func(r rune) bool {
			// Убираем пунктуацию с краёв, КРОМЕ дефиса
			return (unicode.IsPunct(r) && r != '-') || unicode.IsSpace(r)
		})

		// "-" словом не считается
		if processedWord == "-" {
			continue
		}

		// Не учитываем пустые слова (например, только тире)
		if processedWord != "" {
			freq[processedWord]++
		}
	}

	// Создаем слайс пар (слово, частота) для сортировки
	type wordFreq struct {
		word string
		freq int
	}

	wordFreqs := make([]wordFreq, 0, len(freq))
	for word, f := range freq {
		wordFreqs = append(wordFreqs, wordFreq{word, f})
	}

	// Сортируем по частоте (убывание) и лексикографически (возрастание) для одинаковых частот
	sort.Slice(wordFreqs, func(i, j int) bool {
		if wordFreqs[i].freq != wordFreqs[j].freq {
			return wordFreqs[i].freq > wordFreqs[j].freq // по убыванию частоты
		}
		return wordFreqs[i].word < wordFreqs[j].word // по возрастанию лексикографически
	})

	// Берем первые 10 слов или все, если их меньше 10
	result := make([]string, 0, len(wordFreqs))
	for i, wf := range wordFreqs {
		if i >= 10 {
			break
		}
		result = append(result, wf.word)
	}

	return result
}
