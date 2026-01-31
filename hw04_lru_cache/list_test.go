package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
	t.Run("complex", func(t *testing.T) {
		l := NewList()
		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())
		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())
		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]
		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)
		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("push_front_and_back", func(t *testing.T) {
		l := NewList()

		// Тест добавления элемента в пустой список
		item1 := l.PushFront(1)
		require.Equal(t, 1, l.Len())
		require.Equal(t, item1, l.Front())
		require.Equal(t, item1, l.Back())
		require.Equal(t, 1, item1.Value)

		// Проверка добавления элемента в непустой список
		item2 := l.PushBack(2)
		require.Equal(t, 2, l.Len())
		require.Equal(t, item1, l.Front())
		require.Equal(t, item2, l.Back())
		require.Equal(t, 1, item1.Value)
		require.Equal(t, 2, item2.Value)

		// Тест PushFront
		item3 := l.PushFront(3)
		require.Equal(t, 3, l.Len())
		require.Equal(t, item3, l.Front())
		require.Equal(t, item2, l.Back())
		require.Equal(t, 3, item3.Value)
		require.Equal(t, 2, item2.Value)
	})

	t.Run("remove_elements", func(t *testing.T) {
		l := NewList()

		// Проверка удаления из списка, состоящего из одного элемента
		item := l.PushFront(1)
		l.Remove(item)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		// Тест удаления из многоэлементного списка
		item1 := l.PushFront(1)
		item2 := l.PushBack(2)
		item3 := l.PushBack(3)

		// Удалить средний элемент
		l.Remove(item2)
		require.Equal(t, 2, l.Len())
		require.Equal(t, item1, l.Front())
		require.Equal(t, item3, l.Back())

		// Удалить первый элемент
		l.Remove(item1)
		require.Equal(t, 1, l.Len())
		require.Equal(t, item3, l.Front())
		require.Equal(t, item3, l.Back())

		// Удалить последний элемент
		l.Remove(item3)
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("move_to_front", func(t *testing.T) {
		l := NewList()

		// Создать список [1] -> [2] -> [3]
		item1 := l.PushFront(1)
		item2 := l.PushBack(2)
		item3 := l.PushBack(3)

		// Проверить начальное состояние
		require.Equal(t, 3, l.Len())
		require.Equal(t, item1, l.Front())
		require.Equal(t, item3, l.Back())

		// Переместить item2 (value 2) в начало списка
		l.MoveToFront(item2)

		// После перемещения список должен быть [2] -> [1] -> [3]
		require.Equal(t, 3, l.Len())
		require.Equal(t, item2, l.Front()) // item2 should now be front
		require.Equal(t, item3, l.Back())  // item3 should still be back

		// Верификация
		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{2, 1, 3}, elems)
	})

	t.Run("edge_cases", func(t *testing.T) {
		l := NewList()

		// Проверка с нулевыми значениями
		item := l.PushFront(nil)
		require.Equal(t, 1, l.Len())
		require.Equal(t, nil, item.Value)

		// Проведите тестирование с использованием различных типов
		l.PushBack("string")
		l.PushBack(42)
		l.PushBack([]int{1, 2, 3})

		require.Equal(t, 4, l.Len())

		// Проверка удаления всех элементов
		for i := l.Front(); i != nil; {
			next := i.Next
			l.Remove(i)
			i = next
		}
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}
