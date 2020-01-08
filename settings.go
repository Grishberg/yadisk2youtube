package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type History interface {
	IsExists(name string) bool
	AddToHistory(name string)
	Close()
}

type SqlHistory struct {
	db *sql.DB
}

func NewHistory() *SqlHistory {
	database, _ := sql.Open("sqlite3", "settings.db")
	statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS history (id INTEGER PRIMARY KEY, filename TEXT UNIQUE)")

	statement.Exec()
	h := &SqlHistory{database}
	return h
}

func (h *SqlHistory) IsExists(name string) bool {
	rows, _ := h.db.Query("SELECT filename FROM history WHERE filename = $1", name)

	var filename string = ""
	for rows.Next() {
		rows.Scan(&filename)
		return name == filename
	}
	return false
}

func (h *SqlHistory) AddToHistory(name string) {
	statement, _ :=
		h.db.Prepare("INSERT INTO history (filename) VALUES (?)")
	statement.Exec(name)
}

func (h *SqlHistory) Close() {
	h.db.Close()
}
