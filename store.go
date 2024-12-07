package main

import "database/sql"

type Message struct {
	Type   string
	Text   string
	Sender string
}

type Storage struct {
	db *sql.DB
}
