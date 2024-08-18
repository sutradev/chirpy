package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	auth "github.com/sutradev/chirpy/internal/auth"
	database "github.com/sutradev/chirpy/internal/db"
	"golang.org/x/crypto/bcrypt"
)

func (cfg *apiConfig) handlePOSTUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	jsonStruct := database.User{}
	err := decoder.Decode(&jsonStruct)
	if err != nil {
		responseWithError(w, 500, `{"error": "could not decode to User struct"}`)
		return
	}
	returnUser, err := cfg.db.CreateUser(jsonStruct.Email, jsonStruct.Password, cfg.jwtSecret)
	if err != nil {
		responseWithError(w, 500, `{"error": "signing token to New User"}`)
		return
	}
	modifiedUser := database.ResponseUser{
		ID:          returnUser.ID,
		Email:       returnUser.Email,
		Token:       returnUser.Token,
		IsChirpyRed: returnUser.IsChirpyRed,
	}
	jsonReturn, err := json.Marshal(modifiedUser)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, `{"error": "on marshaling user"}`)
		return
	}
	responseWithJson(w, http.StatusCreated, jsonReturn)
}

func (cfg *apiConfig) handlePOSTLogin(w http.ResponseWriter, r *http.Request) {
	loginData := struct {
		ID       int    `json:"id"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginData)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, `{"error": "Invalid requeste payload"}`)
		return
	}

	user, err := cfg.db.GetUserByEmail(loginData.Email)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, `{"error": "User Not Found"}`)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		responseWithError(w, http.StatusUnauthorized, `{"error": "Invalid password"}`)
		return
	}

	signedToken, refreshToken, err := auth.MakeToken(
		cfg.jwtSecret,
		60,
		user.ID,
	)
	if err != nil {
		responseWithError(
			w,
			http.StatusInternalServerError,
			`{"error": "Coud not make token for user"}`,
		)
	}

	err = cfg.db.StoreRefreshToken(user.ID, refreshToken)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusInternalServerError)
	}

	type returnParam struct {
		ID            int    `json:"id"`
		Email         string `json:"email"`
		Token         string `json:"token"`
		Refresh_Token string `json:"refresh_token"`
		IsChirpyRed   bool   `json:"is_chirpy_red"`
	}

	modifiedUser := returnParam{
		ID:            user.ID,
		Email:         user.Email,
		Token:         signedToken,
		Refresh_Token: refreshToken,
		IsChirpyRed:   user.IsChirpyRed,
	}

	jsonReturn, err := json.Marshal(modifiedUser)
	if err != nil {
		responseWithError(
			w,
			http.StatusInternalServerError,
			`{"error": "Failed to marshal response"}`,
		)
		return
	}
	responseWithJson(w, http.StatusOK, jsonReturn)
}

func (cfg *apiConfig) handlePUTUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	jsonStruct := database.User{}
	err := decoder.Decode(&jsonStruct)
	// Extract the token from the Authorization header
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusUnauthorized)
		return
	}
	userID, err := auth.VerifyToken(bearerToken, cfg.jwtSecret)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusUnauthorized)
		return
	}

	intID, err := strconv.Atoi(userID)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusInternalServerError)
		return
	}

	foundUser, err := cfg.db.GetUser(intID)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusInternalServerError)
		return
	}
	// Update the user

	updatedUser, err := cfg.db.UpdateUser(foundUser.ID, jsonStruct.Email, jsonStruct.Password)
	if err != nil {
		http.Error(w, "could not update user", http.StatusInternalServerError)
	}

	// Prepare the modified user response
	modifiedUser := modifiedUser{
		ID:    updatedUser.ID,
		Email: updatedUser.Email,
		Token: updatedUser.AuthData.Token,
	}

	// Write the JSON response
	jsonData, err := json.Marshal(modifiedUser)
	if err != nil {
		http.Error(w, `{"error": "Failed to marshal response"}`, http.StatusInternalServerError)
		return
	}
	responseWithJson(w, http.StatusOK, jsonData)
}

func (cfg *apiConfig) handlePOSTRefresh(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		http.Error(w, "Request body is not allowed", http.StatusUnauthorized)
		return
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusBadRequest)
		return
	}
	user, ok := cfg.db.FindTokenCheckDate(bearerToken)
	if !ok {
		http.Error(w, "User Token Not Found", http.StatusNoContent)
		return
	}
	jwtToken, _, err := auth.MakeToken(cfg.jwtSecret, 60, user.ID)
	if err != nil {
		http.Error(w, "Unable to make new token", http.StatusNoContent)
		return
	}
	type returnToken struct {
		Token string `json:"token"`
	}

	returnT := returnToken{
		Token: jwtToken,
	}
	jsonData, err := json.Marshal(returnT)
	if err != nil {
		http.Error(w, `{"error": "Failed to marshal response"}`, http.StatusInternalServerError)
		return
	}

	responseWithJson(w, 200, jsonData)
}

func (cfg *apiConfig) handlePOSTRevoke(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > 0 {
		http.Error(w, "Request body is not allowed", http.StatusUnauthorized)
		return
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		body := fmt.Sprint(err)
		http.Error(w, body, http.StatusBadRequest)
		return
	}
	user, ok := cfg.db.FindTokenCheckDate(refreshToken)
	if !ok {
		http.Error(w, "User Token Not Found", http.StatusNoContent)
		return
	}

	ok = cfg.db.DeleteRefreshToken(user)
	if !ok {
		http.Error(w, "Could not Revoke Token", http.StatusNoContent)
	}

	w.WriteHeader(204)
}
