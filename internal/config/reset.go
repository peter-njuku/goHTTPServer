package config

import (
	"log"
	"net/http"
)

func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	cfg.FileServerHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset back to 0"))
}
