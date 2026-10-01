package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	type apiConfig struct {
		fileserverHits atomic.Int32
	}
	func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
		cfg.fileserverHits.add()
	}
	server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
	}	
	server.ListenAndServe()
}