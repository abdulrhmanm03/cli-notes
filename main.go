package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // SQLite driver
)

func create_notes_table(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        dir TEXT NOT NULL,
		note TEXT NOT NULL
	);`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}

func get_current_dir() (string, error) {
	return os.Getwd()
}

func add_note(db *sql.DB, dir string, note string) error {
	insertSQL := `INSERT INTO notes (dir, note) VALUES (?, ?)`
	_, err := db.Exec(insertSQL, dir, note)
	if err != nil {
		return err
	}
	return nil
}

var ErrNoRowsDeleted = errors.New("no rows was deleted")

func delete_note(db *sql.DB, dir string, id string) error {
	deleteSQL := `DELETE FROM notes WHERE id = ? AND dir = ?`
	result, err := db.Exec(deleteSQL, id, dir)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNoRowsDeleted
	}
	return nil
}

func get_dir_notes(db *sql.DB, dir string) (map[int]string, error) {
	querySQL := `SELECT id, note FROM notes WHERE dir = ?`
	rows, err := db.Query(querySQL, dir)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make(map[int]string)
	for rows.Next() {
		var id int
		var note string
		err = rows.Scan(&id, &note)
		if err != nil {
			return nil, err
		}
		notes[id] = note
	}
	return notes, nil
}

func list_notes(db *sql.DB, dir string) {
	notes, err := get_dir_notes(db, dir)
	if err != nil {
		log.Fatalf("Failed to query data: %v", err)
	}

	if len(notes) == 0 {
		fmt.Println("No notes in this directory")
	} else {
		for id, note := range notes {
			fmt.Printf("%d: %s\n", id, note)
		}
	}
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}

	dbPath := filepath.Join(homeDir, "dev", "notes", "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	err = create_notes_table(db)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	dir, err := get_current_dir()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		list_notes(db, dir)
		return
	}

	switch args[0] {
	case "add":
		if len(args) < 2 {
			log.Fatalf("Missing note")
		}
		note := strings.Join(args[1:], " ")
		err = add_note(db, dir, note)
		if err != nil {
			log.Fatalf("Failed to add note: %v", err)
		}
		fmt.Println("Note added")
		return

	case "delete":
		if len(args) < 2 {
			log.Fatalf("Missing note id")
		}
		id := args[1]
		err = delete_note(db, dir, id)
		if errors.Is(err, ErrNoRowsDeleted) {
			fmt.Printf("No note with id %s\n", id)
		} else if err != nil {
			log.Fatalf("Failed to delete note: %v", err)
		} else {
			fmt.Println("Note deleted")
		}
		return
	}
}
