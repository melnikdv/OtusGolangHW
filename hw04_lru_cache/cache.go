package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.RWMutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Проверяем, есть ли элемент в кэше
	if item, exists := c.items[key]; exists {
		// Обновляем значение и перемещаем в начало
		item.Value.(*cacheItem).value = value
		c.queue.MoveToFront(item)
		return true
	}

	// Создаем новый элемент
	newItem := &cacheItem{
		key:   key,
		value: value,
	}

	// Добавляем в начало очереди
	queueItem := c.queue.PushFront(newItem)
	c.items[key] = queueItem

	// Проверяем, нужно ли удалять элементы
	if c.queue.Len() > c.capacity {
		// Удаляем последний элемент
		lastItem := c.queue.Back()
		cacheItem := lastItem.Value.(*cacheItem)
		delete(c.items, cacheItem.key)
		c.queue.Remove(lastItem)
	}

	return false
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Перемещаем элемент в начало очереди
	c.queue.MoveToFront(item)

	// Возвращаем значение
	cacheItem := item.Value.(*cacheItem)
	return cacheItem.value, true
}

func (c *lruCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}

// Вспомогательная структура для хранения элементов в очереди.
type cacheItem struct {
	key   Key
	value interface{}
}
