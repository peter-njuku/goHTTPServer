package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/peter-njuku/goHTTPServer/internal/auth"
	"github.com/peter-njuku/goHTTPServer/internal/database"
)

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *ApiConfig) HandlerRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	dbUserWithToken, err := cfg.Db.GetUserFromRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	if dbUserWithToken.RevokedAt.Valid && !dbUserWithToken.RevokedAt.Time.IsZero() {
		respondWithError(w, http.StatusUnauthorized, "Refresh token revoked")
		return
	}

	newAccessToken, err := auth.MakeJWT(dbUserWithToken.ID, cfg.Secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create access token")
		log.Print(err)
		return
	}
	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: newAccessToken,
	})
}

func (cfg *ApiConfig) HandlerRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	currentTime := time.Now().UTC()
	err = cfg.Db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		Token:     refreshTokenString,
		RevokedAt: sql.NullTime{Time: currentTime, Valid: true},
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could nor revoke token")
		log.Print(err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
