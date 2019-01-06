package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gopkg.in/natefinch/lumberjack.v2"

	"ordodox/config"
	"ordodox/database"
	"ordodox/templates"
)

var boards []config.Board

func logger(path string) func(http.Handler) http.Handler {
	l := &lumberjack.Logger{Filename: path, MaxSize: 128, MaxBackups: 10}
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  log.New(l, "", log.LstdFlags|log.LUTC),
		NoColor: true,
	})
}

func secure(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		h.ServeHTTP(w, r)
	})
}

func New(boards_ []config.Board, logpath string) *chi.Mux {
	boards = boards_

	r := chi.NewRouter()
	r.Use(logger(logpath))
	r.Use(secure)
	r.Use(middleware.DefaultCompress)
	initRoutes(r)
	return r
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Index(w, boards)
}

func notFound(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(404)
	templates.Error(w, 404)
}

func board(w http.ResponseWriter, r *http.Request) {
	b := chi.URLParam(r, "b")
	threads, err := database.Board(b)
	if err != nil {
		notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templates.Board(w, struct {
		Board string
		Threads [][]database.Post
	}{b, threads})
	if err != nil {
		panic(err)
	}
}

func thread(w http.ResponseWriter, r *http.Request) {
	b := chi.URLParam(r, "b")
	t := chi.URLParam(r, "t")
	op, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		notFound(w, r)
		return
	}
	posts, err := database.Thread(b, op)
	if err != nil {
		notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templates.Thread(w, struct {
		Board string
		Op int64
		Posts []database.Post
	}{b, op, posts})
	if err != nil {
		panic(err)
	}
}
