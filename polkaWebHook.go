package main

import (
	"encoding/json"
	"net/http"

	"github.com/sutradev/chirpy/internal/auth"
)

type requestParam struct {
	Event string `json:"event"`
	Data  userID `json:"data"`
}

type userID struct {
	UserID int `json:"user_id"`
}

func (cfg *apiConfig) handlePolkaWebHook(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetAPIToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	if token != cfg.polkaApiKey {
		http.Error(w, "Unauthorized", 401)
		return
	}
	expectedRequst := requestParam{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&expectedRequst)
	if err != nil {
		responseWithError(w, 500, `{"error": "could not decode to User struct"}`)
		return
	}
	if expectedRequst.Event != "user.upgraded" {
		responseWithError(w, 204, `{"error": "event not recognized"}`)
		return
	}
	user, err := cfg.db.GetUser(expectedRequst.Data.UserID)
	if err != nil {
		responseWithError(w, 404, `{"error": "could not find user"}`)
		return
	}
	ok := cfg.db.UpgradeRedMember(user)
	if !ok {
		responseWithError(w, 500, `{"error": "unable to upgrade user"}`)
		return
	}
	w.WriteHeader(204)
}
