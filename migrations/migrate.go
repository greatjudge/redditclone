package main

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func migrate(db *sql.DB, migrationsDir string) {
	files := []string{
		"sessions.sql",
		"users.sql",
	}
	for _, filename := range files {
		filepath := path.Join(migrationsDir, filename)
		cont, err := os.ReadFile(filepath)
		if err != nil {
			fmt.Printf("fail to read file %v\n, path: %v\n, %v\n", filename, filepath, err.Error())
			continue
		}
		query := string(cont)
		_, err = db.Exec(query)
		if err != nil {
			fmt.Printf("fail to execute %v: %v\n", filename, err.Error())
			continue
		}
		fmt.Printf("executed for %v\n", filename)
	}
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

// go run migrations/migrate.go
func main() {
	dsn := os.Getenv("MYSQL_DSN")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(1)
	migrate(db, os.Getenv("MIGRATION_DIR"))
}
