package database

import (
	"database/sql"
	"os"
)

// Connect opens postgresql DB
func Connect() (*sql.DB, error) {
	dataSourceName := "host=" + os.Getenv("DB_HOST") + " port=" + os.Getenv("DB_PORT") + " user=" + os.Getenv("DB_USER") + " password=" + os.Getenv("DB_PWD") + " dbname=" + os.Getenv("DB_NAME") + " sslmode=disable"

	db, err := sql.Open("postgres", dataSourceName)

	if err != nil || db.Ping() != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
