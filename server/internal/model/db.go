package model

import (
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

func NewDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}
