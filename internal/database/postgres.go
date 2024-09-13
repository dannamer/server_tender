package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"log"
	"fmt"
	// "tender-service/config"

)

var dbConn *pgx.Conn

func ConnectPostgres() (*pgx.Conn, error) {
	// Получаем строку подключения из переменной окружения
	connStr := "postgresql://your_user:your_password@localhost:5432/your_database"

	// Проверка наличия строки подключения
	if connStr == "" {
		log.Println("Переменная окружения POSTGRES_CONN не установлена")
		return nil, fmt.Errorf("переменная окружения POSTGRES_CONN не установлена")
	}

	// Подключаемся к базе данных
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		// Логируем ошибку, но не завершаем выполнение программы
		log.Printf("Не удалось подключиться к базе данных: %v\n", err)
		return nil, err
	}

	// Успешное подключение
	log.Println("Успешное подключение к базе данных")

	// Сохраняем соединение в глобальную переменную
	dbConn = conn

	return dbConn, nil
}