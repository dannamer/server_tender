package main

import (
	"log"
	"net/http"
	// "tender-service/internal/handlers"
	"tender-service/config"
    // "tender-service/internal/database"
	// "os"
)

func main() {
    // // Подключение к базе данных
    // conn, err := database.ConnectPostgres()
    // if err != nil {
    //     log.Fatalf("Ошибка подключения к базе данных: %v", err)
    // }
    
    // // Создание таблиц в базе данных
    // if err := database.CreatePrivateTables(conn); err != nil {
    //     log.Fatalf("Ошибка создания таблиц: %v", err)
    // }

    // // Привязываем обработчики
    // http.HandleFunc("/api/ping", handlers.PingHandler)
    // http.HandleFunc("/api/tender/new", handlers.CreateTenderHandler)

    // Загружаем конфигурацию
    cfg := config.LoadConfig()
    address := cfg.ServerAddress
    

    log.Println("Server starting on", address)
    log.Println("qwerty", cfg.PostgresConn)
    log.Println("qwerty", cfg.PostgresDB)
    log.Println("qwerty", cfg.PostgresHost)
    log.Println("qwerty", cfg.PostgresJDBCURL)
    log.Println("qwerty", cfg.PostgresPass)
    log.Println("qwerty", cfg.PostgresPort)
    log.Println("qwerty", cfg.PostgresUser)
    log.Println("qwerty", cfg.ServerAddress)
    // Запускаем сервер
    if err := http.ListenAndServe(address, nil); err != nil {
        log.Fatalf("Ошибка при запуске сервера: %v", err)
    }
}

