package main

import (
	"log"
	"net/http"
	"tender-service/internal/handlers"
	"tender-service/config"
	// "os"
)

func main() {
    // Привязываем обработчик для /api/ping
    http.HandleFunc("/api/ping", handlers.PingHandler)

    // Запускаем сервер на 0.0.0.0:8080, чтобы сервер был доступен извне Docker-контейнера
    address := config.LoadConfig().ServerAddress
    log.Println("Server starting on", address)

    if err := http.ListenAndServe(address, nil); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}
