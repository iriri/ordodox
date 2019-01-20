package database

import (
	"sync"

	"ordodox/lru"
)

// i love monomorphizing by hand!!!
type thumbNode struct {
	lru.Node
	thumb []byte
}

var thumbCache struct {
	cache   map[string]*thumbNode
	list    lru.List
	size    int
	maxSize int
	lock    sync.Mutex
}

func initCaches() {
	thumbCache.cache = make(map[string]*thumbNode)
	thumbCache.list.Init()
	thumbCache.maxSize = 128
}

func putCachedThumb(uri string, thumb []byte) {
	thumbCache.lock.Lock()
	n, ok := thumbCache.cache[uri]
	if ok {
		// don't update n.thumb because thumbnails are immutable
		n.Remove()
		thumbCache.list.Push(&n.Node)
		thumbCache.lock.Unlock() // avoid defer's overhead in critical sections
		return
	}

	n = &thumbNode{lru.Node{Key: uri}, thumb}
	thumbCache.list.Push(&n.Node)
	thumbCache.cache[uri] = n
	if thumbCache.size < thumbCache.maxSize {
		thumbCache.size++
	} else {
		delete(thumbCache.cache, thumbCache.list.Shift())
	}
	thumbCache.lock.Unlock()
}

func getCachedThumb(uri string) []byte {
	thumbCache.lock.Lock()
	n, ok := thumbCache.cache[uri]
	if ok {
		n.Remove()
		thumbCache.list.Push(&n.Node)
	}
	thumbCache.lock.Unlock()
	if ok {
		return n.thumb
	}
	return nil
}
