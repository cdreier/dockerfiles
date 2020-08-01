package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cdreier/dockerfiles/singlehostreverseproxy/logger"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {

	target := os.Getenv("TARGET")

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(logger.RequestMiddleware)
	r.Get("/*", getHandler(target))

	log.Println("staring on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Panic(err)
	}

}

func getHandler(target string) http.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		logger.Get().Fatal("unable to start server", err)
	}
	realServer := httputil.NewSingleHostReverseProxy(url)

	return func(w http.ResponseWriter, r *http.Request) {

		// check for routing, append explicit file if none is set
		if filepath.Ext(r.URL.Path) == "" {
			if !strings.HasSuffix(r.URL.Path, "/") {
				r.URL.Path = r.URL.Path + "/"
			}
			r.URL.Path = r.URL.Path + "index.html"
		}

		// Update the headers to allow for SSL redirection
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		// realServer.ServeHTTP(&customResponseWriter{w}, r)
		realServer.ServeHTTP(w, r)
	}
}

type customResponseWriter struct {
	http.ResponseWriter
}

func (w *customResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	log.Println("set code", code)
	if code >= 400 {
		_, err := w.ResponseWriter.Write([]byte("oops"))
		if err != nil {
			log.Println("err", err)
		}
	}
}
