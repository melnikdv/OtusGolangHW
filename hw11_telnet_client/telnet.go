package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// Интерфейс TELNET-клиента.
type TelnetClient interface {
	Connect() error // Устанавливает TCP-соединение с сервером.
	io.Closer       // Закрывает соединение.
	Send() error    // Читает из in и отправляет в сокет.
	Receive() error // Читает из сокета и пишет в out.
}

// Конкретная реализация интерфейса.
type telnetClient struct {
	address string        // Адрес сервера в формате "host:port".
	timeout time.Duration // Таймаут подключения.
	in      io.ReadCloser // Входной поток (обычно os.Stdin).
	out     io.Writer     // Выходной поток (обычно os.Stdout).
	conn    net.Conn      // TCP-соединение.
}

// Создаёт новый TELNET-клиент.
func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

// Устанавливает TCP-соединение с сервером с заданным таймаутом.
func (c *telnetClient) Connect() error {
	var err error
	c.conn, err = net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к %s: %w", c.address, err)
	}
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", c.address)
	return nil
}

// Закрывает соединение и входной поток.
func (c *telnetClient) Close() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil // Обнуляем соединение после закрытия
	}
	if c.in != nil {
		c.in.Close()
	}
	return nil
}

// Читает данные из in и отправляет их в сокет.
func (c *telnetClient) Send() error {
	// Проверяем, что соединение открыто
	if c.conn == nil {
		return fmt.Errorf("соединение закрыто")
	}

	// Копируем всё из in в соединение
	_, err := io.Copy(c.conn, c.in)
	// io.Copy возвращает nil при EOF (например, Ctrl+D), что нормально
	return err
}

// Читает данные из сокета и записывает их в out.
func (c *telnetClient) Receive() error {
	// Проверяем, что соединение открыто
	if c.conn == nil {
		return fmt.Errorf("соединение закрыто")
	}

	// Копируем всё из соединения в out
	_, err := io.Copy(c.out, c.conn)
	// Если сервер закрыл соединение, io.Copy вернёт nil или io.EOF — это штатная ситуация
	return err
}
