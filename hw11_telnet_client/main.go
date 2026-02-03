package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Определяем флаг --timeout с типом duration и значением по умолчанию 10 секунд
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "таймаут подключения к серверу")

	// Парсим аргументы командной строки
	flag.Parse()

	// Проверяем, что переданы ровно два аргумента: хост и порт
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Использование: %s [--timeout=TIME] HOST PORT\n", os.Args[0])
		os.Exit(1)
	}

	// Формируем адрес в виде "host:port"
	address := flag.Arg(0) + ":" + flag.Arg(1)

	// Создаём экземпляр TELNET-клиента
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	// Подключаемся к серверу
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения: %v\n", err)
		os.Exit(1)
	}
	defer client.Close() // Гарантируем закрытие соединения при завершении

	// Создаём контекст, который будет отменён при получении SIGINT (Ctrl+C)
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
	)
	defer cancel()

	// Запускаем горутину для отправки данных из STDIN в сокет
	sendDone := make(chan error, 1)
	go func() {
		sendDone <- client.Send()
	}()

	// Запускаем горутину для получения данных из сокета и вывода в STDOUT
	receiveDone := make(chan error, 1)
	go func() {
		receiveDone <- client.Receive()
	}()

	// Ожидаем завершения одной из операций или сигнала прерывания
	select {
	case <-ctx.Done():
		// Получен SIGINT — завершаем программу
		fmt.Fprintln(os.Stderr, "...Прервано пользователем")
	case err := <-sendDone:
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Ошибка отправки: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
	case err := <-receiveDone:
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Ошибка получения: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
	}
}
