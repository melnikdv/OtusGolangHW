package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// Помогает различать пустые файлы и файлы, у которых первая строка пустая.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// Считывает данные из указанного каталога и возвращает карту переменных окружения.
// Переменные представлены в виде файлов, где filename — это имя переменной, а первая строка файла — это значение.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	env := make(Environment)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("invalid env var name: %q", name)
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		// пустой файл - удалить переменную
		if len(data) == 0 {
			env[name] = EnvValue{
				NeedRemove: true,
			}
			continue
		}

		// занять первую линию
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			data = data[:i]
		}

		// replace \x00 with \n
		data = bytes.ReplaceAll(data, []byte{0x00}, []byte{'\n'})

		// trim spaces and tabs on the right
		value := strings.TrimRight(string(data), " \t")

		env[name] = EnvValue{
			Value:      value,
			NeedRemove: false,
		}
	}

	return env, nil
}
