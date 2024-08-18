package database

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int          `json:"id"`
	Email       string       `json:"email"`
	Password    string       `json:"password"`
	AuthData    RefreshToken `json:"authData"`
	IsChirpyRed bool         `json:"is_chirpy_red"`
}

type RefreshToken struct {
	Token          string    `json:"token"`
	DateMade       time.Time `json:"date_made"`
	ExpirationDate time.Time `json:"expiration_date"`
}

type ResponseUser struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Token       string `json:"token"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func (db *DB) CreateUser(email string, pass string, secret string) (ResponseUser, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return ResponseUser{}, err
	}

	id := len(dbStructure.Users) + 1
	savedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return ResponseUser{}, err
	}

	user := User{
		ID:          id,
		Email:       email,
		Password:    string(savedPass),
		IsChirpyRed: false,
	}

	dbStructure.Users[id] = user

	responseUser := ResponseUser{
		ID:          id,
		Email:       email,
		IsChirpyRed: user.IsChirpyRed,
	}

	err = db.writeDB(dbStructure)
	if err != nil {
		return ResponseUser{}, err
	}

	return responseUser, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, errors.New("User not found")
	}
	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("user not found")
}

func (db *DB) UpdateUser(id int, email string, pass string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	savedPass, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	oldUser, err := db.GetUser(id)
	updatedUser := User{
		ID:       id,
		Email:    email,
		Password: string(savedPass),
		AuthData: RefreshToken{
			Token:          oldUser.AuthData.Token,
			DateMade:       oldUser.AuthData.DateMade,
			ExpirationDate: oldUser.AuthData.ExpirationDate,
		},
		IsChirpyRed: oldUser.IsChirpyRed,
	}

	dbStructure.Users[id] = updatedUser

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return updatedUser, nil
}

func (db *DB) StoreRefreshToken(id int, token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	user, ok := dbStructure.Users[id]
	if !ok {
		return errors.New("User not found")
	}
	updatedUser := User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		AuthData: RefreshToken{
			Token:          token,
			DateMade:       time.Now().UTC(),
			ExpirationDate: time.Now().UTC().Add(time.Hour * 24 * 60),
		},
		IsChirpyRed: user.IsChirpyRed,
	}
	dbStructure.Users[id] = updatedUser
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) FindTokenCheckDate(token string) (User, bool) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, false
	}
	for _, user := range dbStructure.Users {
		if token == user.AuthData.Token {
			return user, true
		}
	}
	return User{}, false
}

func (db *DB) DeleteRefreshToken(user User) bool {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false
	}
	updatedUser := User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		AuthData: RefreshToken{
			Token:          "",
			DateMade:       time.Time{},
			ExpirationDate: time.Time{},
		},
		IsChirpyRed: user.IsChirpyRed,
	}
	dbStructure.Users[user.ID] = updatedUser
	err = db.writeDB(dbStructure)
	if err != nil {
		return false
	}
	return true
}

func (db *DB) UpgradeRedMember(user User) bool {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false
	}
	updatedUser := User{
		ID:       user.ID,
		Email:    user.Email,
		Password: user.Password,
		AuthData: RefreshToken{
			Token:          user.AuthData.Token,
			DateMade:       user.AuthData.DateMade,
			ExpirationDate: user.AuthData.ExpirationDate,
		},
		IsChirpyRed: true,
	}
	dbStructure.Users[user.ID] = updatedUser
	err = db.writeDB(dbStructure)
	if err != nil {
		return false
	}
	return true
}
