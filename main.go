package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/peter-njuku/goHTTPServer/internal/database"
)

const port = "8080"
const filePathRoot = "."

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	apiConfig := ApiConfig{
		Db:       *dbQueries,
		Platform: platform,
		Secret:   secret,
	}

	mux := http.NewServeMux()

	//File Server endpoints
	mux.Handle("/app/", apiConfig.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(filePathRoot)))))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("app/assets"))))

	//Non-fileservers
	mux.HandleFunc("/api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/chirps", apiConfig.handlerValidateChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerChirpsRetrieve)
	mux.HandleFunc("/api/chirps/{chirpID}", apiConfig.handlerChirpGet)
	mux.HandleFunc("POST /api/users", apiConfig.HandlerCreateUser)
	mux.HandleFunc("POST /api/login", apiConfig.handlerLogin)

	//admin endpoints
	mux.HandleFunc("/admin/metrics", apiConfig.HandlerMetrics)
	mux.HandleFunc("/admin/reset", apiConfig.HandlerReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from app on port %s\n", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
