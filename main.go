package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	database "github.com/sutradev/chirpy/internal/db"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	jwtSecret      string
	polkaApiKey    string
}

func main() {
	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaApiKey := os.Getenv("API_KEY")
	const filepathRoot = "."
	const port = "8080"

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}
	cfg := apiConfig{
		fileserverHits: 0,
		db:             db,
		jwtSecret:      jwtSecret,
		polkaApiKey:    polkaApiKey,
	}

	mux := http.NewServeMux()
	mux.Handle(
		"/app/",
		http.StripPrefix(
			"/app",
			cfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot))),
		),
	)

	mux.HandleFunc("GET /admin/metrics", cfg.displayServerHits)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/reset", cfg.resetServerHits)

	mux.HandleFunc("POST /api/chirps", cfg.handlePOSTChirps)
	mux.HandleFunc("GET /api/chirps/", cfg.handleGETValidation)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.handleGetSingleChirp)
	mux.HandleFunc("DELETE /api/chirps/{id}", cfg.handleDeleteChirp)

	mux.HandleFunc("POST /api/users", cfg.handlePOSTUser)
	mux.HandleFunc("PUT /api/users", cfg.handlePUTUser)
	mux.HandleFunc("POST /api/login", cfg.handlePOSTLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlePOSTRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlePOSTRevoke)

	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlePolkaWebHook)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
