package main

import (
	"net/http"
	"chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {

	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	err = cfg.db.RevokeToken(r.Context(), rToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}