package main

import (
	"net/http"
	"chirpy/internal/auth"
	"time"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {

	type response struct {
		Token string `json:"token"`
	}

	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	t, err := cfg.db.GetTokenByID(r.Context(), rToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}
	
	if t.ExpiresAt.Before(time.Now()) || t.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token already expired", nil)
		return
	}

	token, err := auth.MakeJWT(t.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{Token: token,})

}