package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	head   *ListItem
	tail   *ListItem
	length int
}

func NewList() List {
	return &list{}
}

func (l *list) Len() int {
	return l.length
}

func (l *list) Front() *ListItem {
	return l.head
}

func (l *list) Back() *ListItem {
	return l.tail
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := &ListItem{
		Value: v,
		Next:  l.head,
		Prev:  nil,
	}

	if l.head != nil {
		l.head.Prev = newItem
	} else {
		l.tail = newItem
	}

	l.head = newItem
	l.length++

	return newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := &ListItem{
		Value: v,
		Next:  nil,
		Prev:  l.tail,
	}

	if l.tail != nil {
		l.tail.Next = newItem
	} else {
		l.head = newItem
	}

	l.tail = newItem
	l.length++

	return newItem
}

func (l *list) Remove(i *ListItem) {
	if i == nil {
		return
	}

	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.head = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.tail = i.Prev
	}

	l.length--
}

func (l *list) MoveToFront(i *ListItem) {
	if i == nil || i == l.head {
		return
	}

	// Удаляем элемент из текущей позиции
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.head = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.tail = i.Prev
	}

	// Вставляем элемент в начало
	i.Prev = nil
	i.Next = l.head

	if l.head != nil {
		l.head.Prev = i
	} else {
		l.tail = i
	}

	l.head = i
}
