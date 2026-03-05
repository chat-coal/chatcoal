package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

func main() {
	godotenv.Load(".env")

	command := flag.String("cmd", "up", "goose command: up, down, status, reset, version")
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&multiStatements=true",
		os.Getenv("APP_DB_USER"),
		os.Getenv("APP_DB_PASS"),
		os.Getenv("APP_DB_HOST"),
		os.Getenv("APP_DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	switch *command {
	case "up":
		err = goose.Up(db, "migrations")
	case "down":
		err = goose.Down(db, "migrations")
	case "status":
		err = goose.Status(db, "migrations")
	case "reset":
		err = goose.Reset(db, "migrations")
	case "version":
		err = goose.Version(db, "migrations")
	default:
		log.Fatalf("unknown command: %s", *command)
	}

	if err != nil {
		log.Fatalf("goose %s: %v", *command, err)
	}
}
