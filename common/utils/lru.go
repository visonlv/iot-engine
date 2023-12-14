package utils

type lruItem[T any] struct {
	version int64
	key     string
	value   T
	prev    *lruItem[T]
	next    *lruItem[T]
}

// 线程不安全
type LRU[T any] struct {
	size       int
	items      map[string]*lruItem[T]
	dirtyItems map[string]*lruItem[T]
	head       *lruItem[T]
	tail       *lruItem[T]
}

func NewLRU[T any](size int) *LRU[T] {
	lru := &LRU[T]{}
	lru.items = make(map[string]*lruItem[T])
	lru.dirtyItems = make(map[string]*lruItem[T])
	lru.head = new(lruItem[T])
	lru.tail = new(lruItem[T])
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	lru.size = size
	return lru
}

func (lru *LRU[T]) evict() *lruItem[T] {
	item := lru.tail.prev
	lru.pop(item)
	delete(lru.items, item.key)
	return item
}

func (lru *LRU[T]) pop(item *lruItem[T]) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

func (lru *LRU[T]) push(item *lruItem[T]) {
	lru.head.next.prev = item
	item.next = lru.head.next
	item.prev = lru.head
	lru.head.next = item
}

func (lru *LRU[T]) Resize(size int) (evictedItems []T) {
	if size <= 0 {
		panic("invalid size")
	}

	for size < len(lru.items) {
		item := lru.evict()
		evictedItems = append(evictedItems, item.value)
	}
	lru.size = size
	return evictedItems
}

func (lru *LRU[T]) Len() int {
	return len(lru.items)
}

func (lru *LRU[T]) Set(key string, value T, version int64) (evictedKey string, evictedValue T, evicted bool) {
	item := lru.items[key]
	if item == nil {
		if len(lru.items) == lru.size {
			item = lru.evict()
			evictedKey, evictedValue, evicted = item.key, item.value, true
		} else {
			item = new(lruItem[T])
		}
		item.key = key
		item.value = value
		lru.push(item)
		lru.items[key] = item
	} else {
		item.value = value
		if lru.head.next != item {
			lru.pop(item)
			lru.push(item)
		}
	}

	if item.version != 0 {
		lru.dirtyItems[key] = item
	}
	item.version = version

	return evictedKey, evictedValue, evicted
}

func (lru *LRU[T]) Get(key string) (value T, ok bool) {
	item := lru.items[key]
	if item == nil {
		return
	}
	if lru.head.next != item {
		lru.pop(item)
		lru.push(item)
	}
	return item.value, true
}

func (lru *LRU[T]) Contains(key string) bool {
	_, ok := lru.items[key]
	return ok
}

func (lru *LRU[T]) Peek(key string) (value T, ok bool) {
	if item := lru.items[key]; item != nil {
		return item.value, true
	}
	return
}

func (lru *LRU[T]) Delete(key string) (prev T, deleted bool) {
	item := lru.items[key]
	if item == nil {
		return
	}
	delete(lru.items, key)
	lru.pop(item)
	return item.value, true
}

func (lru *LRU[T]) Range(iter func(key string, value T) bool) {
	if head := lru.head; head != nil {
		item := head.next
		for item != lru.tail {
			if !iter(item.key, item.value) {
				return
			}
			item = item.next
		}
	}
}

func (lru *LRU[T]) RangeDirty(iter func(key string, value T) bool) {
	for _, item := range lru.dirtyItems {
		if iter(item.key, item.value) {
			return
		}
	}
}

func (lru *LRU[T]) CleanDirty(key string, version int64) bool {
	if v, ok := lru.dirtyItems[key]; ok {
		if v.version == version {
			delete(lru.dirtyItems, key)
			return true
		}
	}
	return false
}

func (lru *LRU[T]) Clear() {
	lru.size = 0
	lru.items = nil
	lru.head = nil
	lru.tail = nil
}
