package auth

import (
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"github.com/google/uuid"
	"strings"
	"errors"
	"net/http"
	"crypto/rand"
	"encoding/hex"
)

func HashPassword(password string) (string, error) {
    hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
    match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject: userID.String(),
	})

	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var c jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(tokenString, &c, func(token *jwt.Token) (interface{}, error){
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	p, err := uuid.Parse(c.Subject)
	if err != nil {
		return uuid.Nil, err
	}
	return p, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := strings.TrimPrefix(headers.Get("Authorization"), "Bearer ")
	if token == "" {
		return "", errors.New("no authorization found")
	}
	return token, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}