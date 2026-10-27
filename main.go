package main

import (
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"sync/atomic"
	"fmt"
	"os"
	"github.com/joho/godotenv"
	"chirpy/internal/database"
	"database/sql"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
	secret string
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
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment."))
		return
	}

	cfg.fileserverHits.Store(0)
	err := cfg.db.Reset(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state."))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("JWT_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
    	log.Printf("Error conecting: %s", err)
	}
	platform := os.Getenv("PLATFORM")
		if platform == "" {
			log.Fatal("PLATFORM must be set")
	}
	dbQueries := database.New(db)
	cfg := apiConfig{fileserverHits: atomic.Int32{}, db: dbQueries, platform: platform, secret: secret}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.handlerCounter)
	mux.HandleFunc("GET /api/chirps", cfg.handlerChirpsCounter)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerChirpFinder)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", cfg.handlerChirpsCreate)
	mux.HandleFunc("POST /api/users", cfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", cfg.handlerUsersLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevokeToken)
	
	server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
	}	
	log.Fatal(server.ListenAndServe())
}