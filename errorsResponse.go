package main

import "net/http"

func responseWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
