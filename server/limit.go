package server

import (
	"net"
	"net/http"
	"sync"
	"time"
	"unsafe" // fuck it, i tried lmao

	"golang.org/x/time/rate"

	"ordodox/lru"
)

type client struct {
	limiter *rate.Limiter
	seen    time.Time
}

type clientNode struct {
	lru.Node
	client
}

var limiter struct {
	clients map[string]*clientNode
	list    lru.List
	lock    sync.Mutex
}

func cleanupLimiter(t time.Time) {
	for n := limiter.list.Prev; n != &limiter.list; n = n.Prev {
		// yes, this technically violates the unsafe rules but the asm,
		// from 6g at least, looks fine so...
		if t.Sub((*clientNode)(unsafe.Pointer(n)).seen) < 5*time.Minute {
			break
		}
		n.Remove()
		delete(limiter.clients, n.Key)
	}
}

func putLimiter(ip string, l *rate.Limiter) {
	limiter.lock.Lock()
	n, ok := limiter.clients[ip]
	if ok {
		n.seen = time.Now()
		n.Remove()
		limiter.list.Push(&n.Node)
		limiter.lock.Unlock()
		return
	}

	t := time.Now()
	n = &clientNode{lru.Node{Key: ip}, client{limiter: l, seen: t}}
	limiter.list.Push(&n.Node)
	limiter.clients[ip] = n
	cleanupLimiter(t)
	limiter.lock.Unlock()
}

func getLimiter(ip string) *rate.Limiter {
	limiter.lock.Lock()
	n, ok := limiter.clients[ip]
	if ok {
		n.seen = time.Now()
		n.Remove()
		limiter.list.Push(&n.Node)
	}
	limiter.lock.Unlock()
	if ok {
		return n.limiter
	}
	return nil
}

func limit(h http.Handler) http.Handler {
	limiter.clients = make(map[string]*clientNode)
	limiter.list.Init()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		l := getLimiter(ip)
		if l == nil {
			l = rate.NewLimiter(3, 7)
			putLimiter(ip, l)
		}
		if l.Wait(r.Context()) != nil {
			error_(http.StatusTooManyRequests)(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
