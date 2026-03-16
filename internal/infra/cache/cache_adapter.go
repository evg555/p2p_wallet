package cache

import (
	"container/list"
)

var _ List = (*adapterList)(nil)

type adapterList struct {
	lib *list.List
}

func NewList() *adapterList {
	return &adapterList{
		lib: list.New(),
	}
}

func (l *adapterList) Len() int {
	return l.lib.Len()
}

func (l *adapterList) Back() any {
	item := l.lib.Back()
	if item == nil {
		return nil
	}

	return item.Value
}

func (l *adapterList) PushFront(el any) any {
	return l.lib.PushFront(el)
}

func (l *adapterList) Remove(el any) {
	item := l.elementFromAny(el)
	if item == nil {
		return
	}

	l.lib.Remove(item)
}

func (l *adapterList) MoveToFront(el any) {
	item := l.elementFromAny(el)
	if item == nil {
		return
	}

	l.lib.MoveToFront(item)
}

func (l *adapterList) elementFromAny(el any) *list.Element {
	if item, ok := el.(*list.Element); ok {
		return item
	}

	for item := l.lib.Front(); item != nil; item = item.Next() {
		if item.Value == el {
			return item
		}
	}

	return nil
}
