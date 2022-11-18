package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgreSQL struct {
}

func NewPostgreSQL() *PostgreSQL {
	return &PostgreSQL{}
}

func (d *PostgreSQL) ConnectPostgreSQL(host, user, pass, database string, port int) (*sql.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", host, port, user, pass, database)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	return db, nil
}
