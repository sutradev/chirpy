package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++ // Increment the counter on each request
		next.ServeHTTP(w, r) // Call the next handler in the chain
	})
}

func (cfg *apiConfig) displayServerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	d := cfg.fileserverHits
	body := fmt.Sprintf(`
    <html>
        <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %v times!</p>
        </body>
    </html>
    `, d)
	w.Write([]byte(body))
}

func (cfg *apiConfig) resetServerHits(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits = 0
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
