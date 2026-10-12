package main

import (
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"sync/atomic"
	"fmt"
	"encoding/json"
	"strings"
	"os"
	"github.com/joho/godotenv"
	"chirpy/internal/database"
	"database/sql"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, req)
	})
}

func (cfg *apiConfig) handlerCounter(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
    w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(
	`<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
    w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handlerValidate(w http.ResponseWriter, req *http.Request) {
	type parameters struct{
		Body string `json:"body"`
	}

	type returnVals struct{
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(req.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	splitted := strings.Split(params.Body, " ")
	for i, v := range splitted{
		if strings.ToLower(v) == "kerfuffle" || strings.ToLower(v) == "sharbert" || strings.ToLower(v) == "fornax"{
			splitted[i] = "****"
		}
	}
	joined := strings.Join(splitted, " ")

	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: joined,
	})
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
    	log.Printf("Error conecting: %s", err)
	}
	dbQueries := database.New(db)
	cfg := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerCounter)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.handlerValidate)
	
	server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
	}	
	log.Fatal(server.ListenAndServe())
}