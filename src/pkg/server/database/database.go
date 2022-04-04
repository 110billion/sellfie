package database

import "database/sql"

// Connect opens postgresql DB
func Connect() (*sql.DB, error) {
	db, err := sql.Open("postgres", `host=34.64.237.78 port=5432
	user=postgres password=test dbname=postgres sslmode=disable`)

	if err != nil || db.Ping() != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
