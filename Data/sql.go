package Data

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./Data/Data.db")
	if err != nil {
		return nil, err
	}

	// Créez la table utilisateurs si elle n'existe pas
	createTable := `
	CREATE TABLE IF NOT EXISTS utilisateurs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}
	return db, nil
}
