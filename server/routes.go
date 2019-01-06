package server

import (
	"net/http"

	"github.com/go-chi/chi"

	"ordodox/templates"
)

func initRoutes(r *chi.Mux) {
	r.Get("/", index)
	r.Get("/{b}/", board(false))
	r.Get("/{b}/{t}", thread(false))
	r.Get("/css/reset.css", templates.Reset)
	r.Get("/css/ordodox.css", templates.Ordodox)
	r.Get("/json/{b}/", board(true))
	r.Get("/json/{b}/{t}", thread(true))
	r.Post("/{b}/submit", submit)
	r.Post("/{b}/{t}/reply", submit)
	r.NotFound(error_(http.StatusNotFound))
	r.MethodNotAllowed(error_(http.StatusMethodNotAllowed))
}
