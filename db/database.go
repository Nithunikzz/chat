package db

import (
	"database/sql"
	"log"

	//_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"
)

func InitDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, err
	}

	// Initialize tables
	query := `
	CREATE TABLE IF NOT EXISTS clients (
		id VARCHAR(50) PRIMARY KEY,
		active BOOLEAN DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_id VARCHAR(50),
		message TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		return nil, err
	}

	log.Println("Database initialized")
	return db, nil
}
