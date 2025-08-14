package main

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		t.Errorf("could not create database")
	}

	// check if connected
	if err := db.PingContext(context.Background()); err != nil {
		t.Errorf("%s", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		age INTEGER
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		t.Errorf("could not create table: %s", err)
	}
}
