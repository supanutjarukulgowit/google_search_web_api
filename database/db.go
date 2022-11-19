package database

import (
	"database/sql"
	"fmt"

	// _ "github.com/lib/pq"
	"github.com/supanutjarukulgowit/google_search_web_api/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func (d *PostgreSQL) ConnectPostgreSQLGorm(host, user, pass, database string, port int) (*gorm.DB, error) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", host, port, user, pass, database)
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	migrations(db)
	return db, nil
}

func migrations(db *gorm.DB) {
	db.AutoMigrate(
		&model.User{},
		&model.ConfigurationDb{},
		&model.GoogleSearchApiDb{},
		&model.GoogleSearchApiDetailDb{},
	)
}
