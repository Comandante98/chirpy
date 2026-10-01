package main

import (
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
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
    w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + strconv.Itoa(int((cfg.fileserverHits.Load())))))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
    w.WriteHeader(http.StatusOK)
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("/metrics", cfg.handlerCounter)

	mux.HandleFunc("/reset", cfg.handlerReset)
	
	server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
	}	
	log.Fatal(server.ListenAndServe())
}