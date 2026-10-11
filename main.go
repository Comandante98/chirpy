package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"fmt"
	"encoding/json"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
		Valid bool `json:"valid"`
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

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}

func main() {
	cfg := apiConfig{}
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