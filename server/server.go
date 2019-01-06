package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

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

func error_(code int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		if strings.Contains(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templates.Error(w, code)
		}
	}
}

func writeJson(w http.ResponseWriter, v interface{}) {
	j, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)
}

func board(json_ bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := chi.URLParam(r, "b")
		threads, err := database.Board(b)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		if json_ {
			writeJson(w, threads)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err = templates.Board(w, struct {
				Board   string
				Threads [][]database.Post
			}{b, threads})
			if err != nil {
				panic(err)
			}
		}
	}
}

func thread(json_ bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := chi.URLParam(r, "b")
		t := chi.URLParam(r, "t")
		op, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		posts, err := database.Thread(b, op)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		if json_ {
			writeJson(w, posts)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err = templates.Thread(w, struct {
				Board string
				Op    int64
				Posts []database.Post
			}{b, op, posts})
			if err != nil {
				panic(err)
			}
		}
	}
}

func submit(w http.ResponseWriter, r *http.Request) {
	b := chi.URLParam(r, "b")
	t := chi.URLParam(r, "t")
	var op int64
	var err error
	if t == "" {
		op = -1
	} else {
		op, err = strconv.ParseInt(t, 10, 64)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
	}
	var req database.Request
	json_ := strings.Contains(r.Header.Get("Content-Type"), "json")
	if json_ {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			error_(http.StatusBadRequest)(w, r)
			return
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			error_(http.StatusBadRequest)(w, r)
			return
		}
	} else {
		r.ParseForm()
		req.Name = r.PostForm.Get("name")
		req.Email = r.PostForm.Get("email")
		req.Subject = r.PostForm.Get("subject")
		req.Comment = r.PostForm.Get("comment")
	}
	err = database.Submit(b, op, r.RemoteAddr, &req)
	if err != nil {
		error_(http.StatusNotFound)(w, r)
		return
	}
	if json_ {
		w.WriteHeader(http.StatusCreated)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/%s/", b), http.StatusFound)
	}
}
