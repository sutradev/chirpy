package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func MakeToken(secret string, expires int, userID int) (string, string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(
			time.Now().UTC().Add(time.Duration(expires * int(time.Minute))),
		),
		Subject: fmt.Sprintf("%d", userID),
	})
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", "", err
	}
	refreshToken := hex.EncodeToString(bytes)

	return signedToken, refreshToken, nil
}

func VerifyToken(tokenString, tokenSecret string) (string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", err
	}
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", err
	}
	if issuer != string("chirpy") {
		return "", errors.New("invalid issuer")
	}

	return userIDString, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("header not included in request")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}

func GetAPIToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("header not included in request")
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
