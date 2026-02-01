package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetClient_ConnectError(t *testing.T) {
	// Тест проверяет поведение клиента при неудачном подключении к несуществующему серверу
	// Используется неправильный адрес, чтобы вызвать ошибку подключения
	timeout, err := time.ParseDuration("1s")
	require.NoError(t, err)

	client := NewTelnetClient("127.0.0.1:9999", timeout, io.NopCloser(&bytes.Buffer{}), &bytes.Buffer{})
	err = client.Connect()
	require.Error(t, err)
	// Ожидается, что будет ошибка подключения
}

func TestTelnetClient_SendError(t *testing.T) {
	// Тест проверяет поведение при попытке отправки данных в закрытое соединение
	// Создаём клиент и подключаемся к локальному серверу
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { require.NoError(t, l.Close()) }()

	timeout, err := time.ParseDuration("10s")
	require.NoError(t, err)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
	require.NoError(t, client.Connect())

	// Закрываем соединение вручную
	client.Close()

	// Пытаемся отправить данные — это должно привести к ошибке
	err = client.Send()
	require.Error(t, err)
	// Ожидается, что будет ошибка отправки
}

func TestTelnetClient_ReceiveError(t *testing.T) {
	// Тест проверяет поведение при получении данных из закрытого соединения
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { require.NoError(t, l.Close()) }()

	timeout, err := time.ParseDuration("10s")
	require.NoError(t, err)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
	require.NoError(t, client.Connect())

	// Закрываем соединение вручную
	client.Close()

	// Пытаемся получить данные — это должно привести к ошибке
	err = client.Receive()
	require.Error(t, err)
	// Ожидается, что будет ошибка получения
}

func TestTelnetClient_Close(t *testing.T) {
	// Тест проверяет корректность закрытия соединения
	l, err := net.Listen("tcp", "127.0.0.1:")
	require.NoError(t, err)
	defer func() { require.NoError(t, l.Close()) }()

	timeout, err := time.ParseDuration("10s")
	require.NoError(t, err)

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
	require.NoError(t, client.Connect())

	// Вызываем Close и проверяем, что ошибок нет
	err = client.Close()
	require.NoError(t, err)
}
