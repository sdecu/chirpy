package database

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exits

func NewDB(path string) (*DB, error) {
    db := &DB{
        path: path,
        mux:  &sync.RWMutex{},
    }
    
    if err := db.ensureDB(); err != nil {
        return nil, err
    }

    return db, nil
}

func (db *DB) Create	{
	
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	ch := make([]Chirp, 0)
	data, err := os.ReadFile(db.path)
	if err != nil {
		return ch, err
	}

	var dbStruct DBStructure
	if err := json.Unmarshal(data, &dbStruct); err != nil {
		return ch, err
	}

	for _, chirp := range dbStruct.Chirps {
		ch = append(ch, chirp)
	}

	sort.Slice(ch, func(i, j int) bool {
		return ch[i].ID < ch[j].ID
	})

	return ch, nil
}

func (db *DB) ensureDB() error {
    // Check if the file exists
    if _, err := os.Stat(db.path); err != nil {
        if os.IsNotExist(err) {
            // Create the file with initial structure
            initial := DBStructure{Chirps: map[int]Chirp{}}
            data, err := json.Marshal(initial)
            if err != nil {
                return err
            }
            return os.WriteFile(db.path, data, 0644) // Create with initial structure
        }
        return err
    }
    return nil
}
