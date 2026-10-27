package main

import (
	"net/http"
	"encoding/json"
	"chirpy/internal/auth"
	"time"
	"chirpy/internal/database"
)

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	validPassword, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}
	if validPassword == false {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", err)
		return
	}

	rToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: auth.MakeRefreshToken(), UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour * 24 * 60)})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}



	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:               user.ID,
			CreatedAt:        user.CreatedAt,
			UpdatedAt:        user.UpdatedAt,
			Email:            user.Email,
		},
		Token: token,
		RefreshToken: rToken.Token,
	})
}