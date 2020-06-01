package main

import (
	"extendedwebserver/logger"
	"net/http"

	"github.com/cdreier/golang-snippets/snippets"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(logger.RequestMiddleware)

	snippets.ChiFileServer(r, "/", http.Dir("/www"))

	if err := http.ListenAndServe(":80", r); err != nil {
		logger.Get().Fatalw("unable to start webserver", "error", err.Error())
	}
}
