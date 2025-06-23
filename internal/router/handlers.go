package router

import (
	"log"
	"net/http"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("This endpoint was hit: %s "+
		"By this method: %s", r.URL.Path, r.Method)
	w.Write([]byte("The root endpoint was hit: " + r.URL.Path))
}
