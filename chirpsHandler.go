package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/sutradev/chirpy/internal/auth"
	database "github.com/sutradev/chirpy/internal/db"
)

func (cfg *apiConfig) handlePOSTChirps(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return
	}
	userID, err := auth.VerifyToken(token, cfg.jwtSecret)
	if err != nil {
		body := fmt.Sprintln(err)
		http.Error(w, body, http.StatusUnauthorized)
		return
	}
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		body := fmt.Sprintln(err)
		http.Error(w, body, http.StatusInternalServerError)
		return
	}

	decoder := json.NewDecoder(r.Body)
	jsonStruct := database.Chirp{}
	err = decoder.Decode(&jsonStruct)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	filteredJson, err := filteredBody(jsonStruct)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		errorMessage := fmt.Sprintln(err)
		w.Write([]byte(errorMessage))
		return
	}
	returnChirp, err := cfg.db.CreateChirp(filteredJson.Body, userIDint)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(201)
	type dataReturn struct {
		ID       int    `json:"id"`
		Body     string `json:"body"`
		AuthorID int    `json:"author_id"`
	}
	finalReturn := dataReturn{
		ID:       returnChirp.ID,
		Body:     returnChirp.Body,
		AuthorID: returnChirp.AuthorID,
	}
	jsonReturn, err := json.Marshal(finalReturn)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.Write(jsonReturn)
}

func filteredBody(jsonStruct database.Chirp) (database.Chirp, error) {
	if len(jsonStruct.Body) > 140 {
		return database.Chirp{}, errors.New(`{"error": "Chirp is too long"}`)
	}

	splicedSentence := strings.Split(jsonStruct.Body, " ")
	for i, bword := range splicedSentence {
		word := strings.ToLower(bword)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			splicedSentence[i] = "****"
		}
	}
	joinedSentence := strings.Join(splicedSentence, " ")
	jsonStruct.Body = joinedSentence
	return jsonStruct, nil
}

func (cfg *apiConfig) handleGETValidation(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	defSort := "asc"
	reqSort := r.URL.Query().Get("sort")
	if reqSort == "" {
		reqSort = defSort
	}
	if s != "" {
		id, err := strconv.Atoi(s)
		if err != nil {
			http.Error(w, "could not convert string id", 500)
			return
		}
		authorChirps, err := cfg.db.GetAuthorChirps(id)
		if err != nil {
			http.Error(w, "could not get chirps for author", 500)
			return
		}
		if reqSort == "desc" {
			sort.Slice(authorChirps, func(i, j int) bool {
				return authorChirps[i].ID > authorChirps[j].ID
			})
		} else {
			sort.Slice(authorChirps, func(i, j int) bool {
				return authorChirps[i].ID < authorChirps[j].ID
			})
		}
		data, err := json.Marshal(authorChirps)
		if err != nil {
			http.Error(w, "Could not marshal data", 500)
			return
		}
		responseWithJson(w, 200, data)
		return
	}
	chirps, err := cfg.db.GetChirps()
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	if reqSort == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	jsonChirps, err := json.Marshal(chirps)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}

	w.Write(jsonChirps)
}

func (cfg *apiConfig) handleGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("id")
	chirpIdInt, err := strconv.Atoi(chirpID)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	chirp, err := cfg.db.GetSingleChirp(chirpIdInt)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	jsonChirp, err := json.Marshal(chirp)
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}
	w.Write(jsonChirp)
}

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return
	}
	userID, err := auth.VerifyToken(token, cfg.jwtSecret)
	if err != nil {
		body := fmt.Sprintln(err)
		http.Error(w, body, http.StatusUnauthorized)
		return
	}
	userIDint, err := strconv.Atoi(userID)
	chirpID := r.PathValue("id")
	chirpIDInt, err := strconv.Atoi(chirpID)
	if err != nil {
		http.Error(w, "error on converstion", http.StatusBadRequest)
		return
	}

	chirp, err := cfg.db.GetSingleChirp(chirpIDInt)
	if err != nil {
		http.Error(w, "error on getting chirp", http.StatusBadRequest)
		return
	}

	if userIDint != chirp.AuthorID {
		http.Error(w, "Not correct of Author of chirp", 403)
		return
	}
	err = cfg.db.DeleteChirp(chirpIDInt)
	if err != nil {
		http.Error(w, "Unable to Delete chirp", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(204)
}
