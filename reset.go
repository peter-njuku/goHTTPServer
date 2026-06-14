package main

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

	if cfg.Platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden: Reset only allowed in development world!")
		return
	}

	err := cfg.Db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to reset database")
		log.Print(err)
		return
	}

	cfg.FileServerHits.Store(0)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset back to 0"))
}
