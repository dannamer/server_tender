package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"
	"os"
)

var dbConn *pgx.Conn

func ConnectPostgres() {
	var err error
	dbConn, err = pgx.Connect(context.Background(), os.Getenv("POSTGRES_CONN"))
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v\n", err)
	}
	log.Println("Успешное подключение к базе данных")
}