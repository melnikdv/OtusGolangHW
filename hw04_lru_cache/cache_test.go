package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", testEmptyCache)
	t.Run("simple", testSimple)
	t.Run("purge logic", testPurgeLogic)
	t.Run("get and set same key", testGetAndSetSameKey)
	t.Run("clear cache", testClearCache)
	t.Run("lru behavior", testLRUBehavior)
}

func testEmptyCache(t *testing.T) {
	c := NewCache(10)

	_, ok := c.Get("aaa")
	require.False(t, ok)

	_, ok = c.Get("bbb")
	require.False(t, ok)
}

func testSimple(t *testing.T) {
	c := NewCache(5)
	wasInCache := c.Set("aaa", 100)
	require.False(t, wasInCache)
	wasInCache = c.Set("bbb", 200)
	require.False(t, wasInCache)
	val, ok := c.Get("aaa")
	require.True(t, ok)
	require.Equal(t, 100, val)
	val, ok = c.Get("bbb")
	require.True(t, ok)
	require.Equal(t, 200, val)
	wasInCache = c.Set("aaa", 300)
	require.True(t, wasInCache)
	val, ok = c.Get("aaa")
	require.True(t, ok)
	require.Equal(t, 300, val)
	val, ok = c.Get("ccc")
	require.False(t, ok)
	require.Nil(t, val)
}

func testPurgeLogic(t *testing.T) {
	// Проверка лимита вместимости
	c := NewCache(3)

	// Заполенение кеша
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Проверьте наличие всех элементов
	val, ok := c.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, val)

	val, ok = c.Get("b")
	require.True(t, ok)
	require.Equal(t, 2, val)

	val, ok = c.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, val)

	// Добавить еще один элемент — следует удалить наименее недавно использованный (элемент 'а")
	c.Set("d", 4)

	// Проверка, что "а" был выселен
	_, ok = c.Get("a")
	require.False(t, ok)

	// Убедитесь, что "b" и "c" по-прежнему присутствуют
	val, ok = c.Get("b")
	require.True(t, ok)
	require.Equal(t, 2, val)

	val, ok = c.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, val)

	// Убедитесь, что "d" присутствует.
	val, ok = c.Get("d")
	require.True(t, ok)
	require.Equal(t, 4, val)

	// Тестовое обновление существующего элемента перемещает его вперед
	c.Set("b", 20) // Обновить элемент "b"
	c.Set("e", 5)  // Добавить новый элемент - следует удалить "c"

	// "c" следует исключить (наименее недавно использованный)
	_, ok = c.Get("c")
	require.False(t, ok)

	// Буква "b" должна оставаться на месте (она была обновлена, поэтому является более актуальной)
	val, ok = c.Get("b")
	require.True(t, ok)
	require.Equal(t, 20, val)

	// "e" должно присутствовать
	val, ok = c.Get("e")
	require.True(t, ok)
	require.Equal(t, 5, val)
}

func testGetAndSetSameKey(t *testing.T) {
	c := NewCache(3)

	// Установить ключ
	c.Set("key1", "value1")

	// Верните это обратно
	val, ok := c.Get("key1")
	require.True(t, ok)
	require.Equal(t, "value1", val)

	// Установите значение еще раз, но с другим значением.
	c.Set("key1", "value2")

	// Получите обновленное значение
	val, ok = c.Get("key1")
	require.True(t, ok)
	require.Equal(t, "value2", val)
}

func testClearCache(t *testing.T) {
	c := NewCache(3)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Verify items are present
	_, ok := c.Get("a")
	require.True(t, ok)

	_, ok = c.Get("b")
	require.True(t, ok)

	_, ok = c.Get("c")
	require.True(t, ok)

	// Очистка кэша
	c.Clear()

	// Убедитесь, что все предметы убраны
	_, ok = c.Get("a")
	require.False(t, ok)

	_, ok = c.Get("b")
	require.False(t, ok)

	_, ok = c.Get("c")
	require.False(t, ok)

	// Проверка, что кэш пуст
	require.Equal(t, 0, c.(*lruCache).queue.Len())
	require.Equal(t, 0, len(c.(*lruCache).items))
}

func testLRUBehavior(t *testing.T) {
	c := NewCache(3)

	// Add 3 items
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// "a", чтобы выбрать последний использованный элемент
	c.Get("a")

	// Добавить 4-й элемент - "d"
	// "b" следует удалить (последний использованный)
	c.Set("d", 4)

	// Проверка, что "b" был выселен
	_, ok := c.Get("b")
	require.False(t, ok)

	// Check that "a", "c", "d" are still there
	val, ok := c.Get("a")
	require.True(t, ok)
	require.Equal(t, 1, val)

	val, ok = c.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, val)

	val, ok = c.Get("d")
	require.True(t, ok)
	require.Equal(t, 4, val)

	// "c", чтобы выбрать ее как последнюю использованную
	c.Get("c")

	// Добавить еще один элемент - "e"
	// "a" следует удалить (последний использованный из оставшихся)
	c.Set("e", 5)

	// Проверка, что "а" был выселен
	_, ok = c.Get("a")
	require.False(t, ok)

	// Проверка, что символы "c", "d", "e" по-прежнему присутствуют
	val, ok = c.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, val)

	val, ok = c.Get("d")
	require.True(t, ok)
	require.Equal(t, 4, val)

	val, ok = c.Get("e")
	require.True(t, ok)
	require.Equal(t, 5, val)
}

func TestCacheMultithreading(_ *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
}
