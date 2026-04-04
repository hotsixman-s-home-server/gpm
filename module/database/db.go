package database

import (
	"database/sql"
	"gpm/module/util"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type _DB struct {
	db *sql.DB
}

var DB *_DB

func init() {
	var err error = nil
	DB, err = OpenDB()
	if err != nil {
		log.Println("Cannot open database.")
		os.Exit(1)
	}
}

func (this _DB) Close() error {
	return this.db.Close()
}

func (this _DB) UpdateMainLogFile(filename string) error {
	_, err := this.db.Exec("INSERT OR REPLACE `logfile-name` (filename) VALUES (?)", filename)
	return err
}

func (this _DB) UpdateLogFile(processName string, filename string) error {
	_, err := this.db.Exec("INSERT OR REPLACE `logfile` (name, filename) VALUES (?, ?)", processName, filename)
	return err
}

var initQueries []string = []string{
	"CREATE TABLE IF NOT EXISTS pid (pid INTEGER);",
	"CREATE TABLE IF NOT EXISTS logfile (name TEXT UNIQUE, filename TEXT);",
	`CREATE TABLE IF NOT EXISTS "logfile-main" (filename TEXT);`,
}

func OpenDB() (*_DB, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	for _, query := range initQueries {
		_, err = db.Exec(query)
		if err != nil {
			return nil, err
		}
	}

	return &_DB{db}, nil
}

func getDBPath() (string, error) {
	homeDir, err := util.GetHomeDirPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "main.db"), nil
}
