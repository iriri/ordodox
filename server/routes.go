package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"ordodox/static"
)

func initRoutes(mux *chi.Mux) {
	mux.Get("/", index)
	mux.Get("/{b}", redirect)
	mux.Get("/{b}/", board(false))
	mux.Get("/{b}/{t}", thread(false))
	mux.Get("/img/{i}", image)
	mux.Get("/thumb/{t}", thumb)
	mux.Get("/css/ordodox.css", static.Ordodox)
	mux.Get("/json/{b}/", board(true))
	mux.Get("/json/{b}/{t}", thread(true))

	mux.Post("/{b}/submit", submit)
	mux.Post("/{b}/{t}/reply", reply)

	mux.Get("/robots.txt", static.Robots)
	mux.Get("/favicon.ico", static.Favicon)

	mux.NotFound(error_(http.StatusNotFound))
	mux.MethodNotAllowed(error_(http.StatusMethodNotAllowed))
}
