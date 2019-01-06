package server

import (
	"github.com/go-chi/chi"

	"ordodox/templates"
)

func initRoutes(r *chi.Mux) {
	r.Get("/", index)
	r.Get("/{b}/", board)
	r.Get("/{b}/{t}", thread)
	r.Get("/css/reset.css", templates.Reset)
	r.Get("/css/ordodox.css", templates.Ordodox)
	r.NotFound(notFound)
}
