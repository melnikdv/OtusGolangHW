package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// Выполняет команду + аргументы (cmd) с использованием переменных окружения из env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	// #nosec G204 -- Выполнение произвольной команды является запланированным поведением.
	command := exec.Command(cmd[0], cmd[1:]...)

	// stdio passthrough
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	// base environment
	envMap := make(map[string]string)
	for _, e := range os.Environ() {
		if p := indexRune(e, '='); p >= 0 {
			envMap[e[:p]] = e[p+1:]
		}
	}

	for k, v := range env {
		if v.NeedRemove {
			delete(envMap, k)
		} else {
			envMap[k] = v.Value
		}
	}

	finalEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		finalEnv = append(finalEnv, k+"="+v)
	}

	command.Env = finalEnv

	err := command.Run()
	if err == nil {
		return 0
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}

	return 1
}

// Вспомогательный инструмент, позволяющий избежать импорта строк.
func indexRune(s string, r rune) int {
	for i, c := range s {
		if c == r {
			return i
		}
	}
	return -1
}
