package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"insider-go-project/internal/message-processor/config"
	"log"
	"time"
)

const insertStatement = `INSERT INTO messages (msisdn, content, sent) VALUES ($1, $2, $3)`

func main() {
	var count int

	flag.IntVar(&count, "n", 100, "number of fake messages to generate & insert")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: no .env file loaded")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("loading configs", err)
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatalf("opening db: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("closing db: %v", err)
		}
	}(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	gofakeit.Seed(time.Now().UnixNano())

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}

	stmt, err := tx.PrepareContext(ctx, insertStatement)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("prepare stmt: %v", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			log.Fatalf("closing db: %v", err)
		}
	}(stmt)

	for i := 0; i < count; i++ {
		msisdn := fakeMSISDN()
		content := gofakeit.Sentence(12) // typical SMS length
		if len(content) > 160 {
			content = content[:160]
		}
		sent := gofakeit.Bool()

		if _, err := stmt.ExecContext(ctx, msisdn, content, sent); err != nil {
			_ = tx.Rollback()
			log.Fatalf("insert row %d: %v", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit tx: %v", err)
	}

	fmt.Printf("Successfully inserted %d fake messages.\n", count)
}

func fakeMSISDN() string {
	num := gofakeit.PhoneFormatted()
	digits := make([]rune, 0, 15)
	for _, r := range num {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
		}
	}
	if len(digits) > 15 {
		digits = digits[len(digits)-15:]
	}
	return string(digits)
}
