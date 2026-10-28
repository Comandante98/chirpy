package main

import (
	"net/http"
	"github.com/google/uuid"
	"chirpy/internal/database"
)

func authorIDFromRequest(r *http.Request) (uuid.UUID, error) {
	authorIDString := r.URL.Query().Get("author_id")
	if authorIDString == "" {
		return uuid.Nil, nil
	}
	authorID, err := uuid.Parse(authorIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return authorID, nil
}

func (cfg *apiConfig) handlerChirpsCounter(w http.ResponseWriter, r *http.Request) {
	authorID, err := authorIDFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
		return
	}

	var dbChirps []database.Chirp
	var err2 error

	if authorID != uuid.Nil {
		dbChirps, err2 = cfg.db.GetChirpsByAuthor(r.Context(), authorID)
	} else {
		dbChirps, err2 = cfg.db.GetChirps(r.Context())
	}
	if err2 != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err2)
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}