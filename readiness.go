package main

import (
	"log"
	"net/http"
)

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
