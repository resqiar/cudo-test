package configs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDBCon() *pgxpool.Pool {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	DSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := pgxpool.New(context.Background(), DSN)
	if err != nil {
		log.Fatalf("Error initializing db connection: %s", err.Error())
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatalf("Error ping-ing db: %s", err.Error())
	}

	return db
}
