package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	s := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Write([]byte(s))
}

func main() {
	const port = "8080"
	const filePathRoot = "."

	apiConfig := &apiConfig{}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiConfig.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("app/assets"))))

	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("/admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("/admin/reset", apiConfig.handlerReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from app on port %s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
