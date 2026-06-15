package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/peter-njuku/goHTTPServer/internal/auth"
)

type PolkaWebhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

func (cfg *ApiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	providedKey, err := auth.GetAPIKey(r.Header)
	if err != nil || providedKey != cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, "Missing API Key")
	}

	decoder := json.NewDecoder(r.Body)
	params := PolkaWebhookRequest{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding request")
		log.Print(err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.Db.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found")
		log.Print(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
