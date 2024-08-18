package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorID: authorID,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) DeleteChirp(id int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	for _, chirp := range dbStructure.Chirps {
		if chirp.ID == id {
			delete(dbStructure.Chirps, id)
			return nil
		}
	}
	return errors.New("did not find chirp")
}

func (db *DB) GetSingleChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	singleChirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, errors.New("chirp not found")
	}
	return singleChirp, nil
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetAuthorChirps(authorID int) ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0)
	for _, chirp := range dbStructure.Chirps {
		if chirp.AuthorID == authorID {
			chirps = append(chirps, chirp)
		}
	}

	return chirps, nil
}
