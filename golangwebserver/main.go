package main

import (
	"net/http"
)

func main() {

	handler := http.FileServer(http.Dir("/www"))
	http.Handle("/", handler)

	http.ListenAndServe(":80", nil)
}
