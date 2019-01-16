package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"ordodox/static"
)

func initRoutes(r *chi.Mux) {
	r.Get("/", index)
	r.Get("/{b}", redirect)
	r.Get("/{b}/", board(false))
	r.Get("/{b}/{t}", thread(false))
	r.Get("/img/{i}", image)
	r.Get("/thumb/{t}", thumb)
	r.Get("/css/reset.css", static.Reset)
	r.Get("/css/ordodox.css", static.Ordodox)
	r.Get("/json/{b}/", board(true))
	r.Get("/json/{b}/{t}", thread(true))
	r.Post("/{b}/submit", submit)
	r.Post("/{b}/{t}/reply", submit)
	r.NotFound(error_(http.StatusNotFound))
	r.MethodNotAllowed(error_(http.StatusMethodNotAllowed))
}
