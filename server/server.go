package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"regexp"
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

// don't think i need this since i'm limiting manually when parsing post requests?
/*
func limit(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 0x400000)
		h.ServeHTTP(w, r)
	})
}
*/

func secure(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		h.ServeHTTP(w, r)
	})
}

func New(boards_ []config.Board, logpath string) http.Handler {
	boards = boards_

	r := chi.NewRouter()
	r.Use(logger(logpath))
	// r.Use(limit)
	r.Use(secure)
	r.Use(middleware.DefaultCompress)
	initRoutes(r)
	return r
}

func index(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Index(w, boards)
}

// TODO: actually log stuff
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

func board(isJson bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := chi.URLParam(r, "b")
		threads, err := database.GetBoard(b)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		if isJson {
			writeJson(w, threads)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err = templates.Board(w, struct {
				Board   string
				Threads [][]*database.Post
			}{b, threads})
			if err != nil {
				panic(err)
			}
		}
	}
}

func thread(isJson bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b := chi.URLParam(r, "b")
		t := chi.URLParam(r, "t")
		op, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		posts, err := database.GetThread(b, op)
		if err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
		if isJson {
			writeJson(w, posts)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err = templates.Thread(w, struct {
				Board string
				Op    int64
				Posts []*database.Post
			}{b, op, posts})
			if err != nil {
				panic(err)
			}
		}
	}
}

func image(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "i")
	img, err := database.GetImage(uri)
	if err != nil {
		error_(http.StatusNotFound)(w, r)
		return
	}
	w.Header().Set("Content-Type", "")
	w.Write(img)
}

func thumb(w http.ResponseWriter, r *http.Request) {
	uri := chi.URLParam(r, "t")
	img, err := database.GetThumb(uri)
	if err != nil {
		error_(http.StatusNotFound)(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(img)
}

type unexpectedField string

func (u unexpectedField) Error() string {
	return fmt.Sprintf("Unexpected field: %s", u)
}

type fieldTooLong struct {
	name string
	n    int
}

func (f fieldTooLong) Error() string {
	return fmt.Sprintf("Field %s exceeded maximum length %d", f.name, f.n)
}

type unsupportedMimetype struct{ name, typ string }

func (u unsupportedMimetype) Error() string {
	return fmt.Sprintf("Unsupported MIME type %s in field %s", u.typ, u.name)
}

type limitReader struct {
	r io.Reader
	n int
}

func (l *limitReader) Read(buf []byte) (int, error) {
	if l.n <= 0 {
		return 0, io.ErrShortBuffer
	}
	if len(buf) > l.n {
		buf = buf[:l.n]
	}
	n, err := l.r.Read(buf)
	l.n -= n
	if l.n == 0 && err == nil {
		return n, io.ErrShortBuffer
	}
	return n, err
}

// MaxBytesReader and ParseMultipartForm are for losers. nih ganggg
func field(form *multipart.Reader, name string, l int) ([]byte, string, error) {
	part, err := form.NextPart()
	if err != nil {
		return nil, "", err
	}
	defer part.Close()
	if part.FormName() != name {
		return nil, "", unexpectedField(name)
	}
	b, err := ioutil.ReadAll(&limitReader{part, l})
	if err != nil {
		if err == io.ErrShortBuffer {
			return nil, "", fieldTooLong{name, l}
		}
		return nil, "", err
	}
	return b, part.FileName(), nil
}

func parseText(form *multipart.Reader, name string, l int) (interface{}, error) {
	b, _, err := field(form, name, l)
	if len(b) == 0 {
		return nil, err
	}
	return string(b), nil
}

var imageTypes = regexp.MustCompile("image/(gif|jpeg|png)")

func parseImage(form *multipart.Reader, name string, l int) ([]byte, string, error) {
	b, fname, err := field(form, name, l)
	if len(b) == 0 {
		return nil, "", err
	}
	if typ := http.DetectContentType(b); !imageTypes.MatchString(typ) {
		return nil, "", unsupportedMimetype{name, typ}
	}
	return b, fname, nil
}

func parseMulti(w http.ResponseWriter, r *http.Request) (*database.Request, error) {
	req := new(database.Request)
	form, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}
	if req.Name, err = parseText(form, "name", 64); err != nil {
		return nil, err
	}
	if req.Email, err = parseText(form, "email", 128); err != nil {
		return nil, err
	}
	if req.Subject, err = parseText(form, "subject", 128); err != nil {
		return nil, err
	}
	if req.Comment, err = parseText(form, "comment", 65536); err != nil {
		return nil, err
	}
	if req.Image, req.ImageName, err = parseImage(form, "image", 0x300000); err != nil {
		return nil, err
	}
	return req, nil
}

func submit(w http.ResponseWriter, r *http.Request) {
	b := chi.URLParam(r, "b")
	t := chi.URLParam(r, "t")
	var op int64
	var err error
	if t == "" {
		op = -1
	} else {
		if op, err = strconv.ParseInt(t, 10, 64); err != nil {
			error_(http.StatusNotFound)(w, r)
			return
		}
	}

	var req *database.Request
	json_ := strings.Contains(r.Header.Get("Content-Type"), "json")
	if json_ { // TODO: what about images?
		body, err := ioutil.ReadAll(&limitReader{r.Body, 0x400000})
		if err != nil {
			error_(http.StatusBadRequest)(w, r)
			return
		}
		req = new(database.Request)
		err = json.Unmarshal(body, req)
	} else {
		req, err = parseMulti(w, r)
	}
	if err != nil {
		error_(http.StatusBadRequest)(w, r)
		return
	}
	if err = database.Submit(b, op, r.RemoteAddr, req); err != nil {
		// TODO: handle different error types
		error_(http.StatusBadRequest)(w, r)
		return
	}
	if json_ {
		w.WriteHeader(http.StatusCreated) // idk what you're actually supposed to do lmao
	} else if strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Redirect(w, r, fmt.Sprintf("/%s/", b), http.StatusFound)
	}
}
