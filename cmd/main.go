package main

import (
	"log"
	"net/http"
	// "tender-service/config"
	"tender-service/internal/handlers"
	"tender-service/internal/database"
)

func main() {
	// Подключение к базе данных
	_, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Привязываем обработчики для API
	http.HandleFunc("/api/ping", handlers.PingHandler)                    // Проверка доступности сервера
	http.HandleFunc("/api/tender/new", handlers.CreateTenderHandler)       // Создание нового тендера
	http.HandleFunc("/api/tender/{tenderId}/status", handlers.GetTenderStatusHandler) // Получение статуса тендера
	http.HandleFunc("/api/tenders", handlers.GetTendersHandler)            // Получение списка тендеров
	http.HandleFunc("/api/tenders/my", handlers.GetUserTendersHandler)     // Получение тендеров пользователя
	http.HandleFunc("/api/tender/{tenderId}/edit", handlers.EditTenderHandler) // Редактирование тендера
	http.HandleFunc("/api/tender/{tenderId}/rollback/{version}", handlers.RollbackTenderHandler) // Откат тендера к версии
	// http.HandleFunc("/api/bids/new", handlers.CreateBidHandler)            // Создание нового предложения
	// http.HandleFunc("/api/bids/my", handlers.GetUserBidsHandler)           // Получение предложений пользователя
	// http.HandleFunc("/api/bids/{tenderId}/list", handlers.GetBidsForTenderHandler) // Получение предложений для тендера
	// http.HandleFunc("/api/bids/{bidId}/status", handlers.GetBidStatusHandler)  // Получение статуса предложения
	// http.HandleFunc("/api/bids/{bidId}/submit_decision", handlers.SubmitBidDecisionHandler) // Отправка решения по предложению

	address := "0.0.0.0:8080"
	log.Printf("Сервер запущен по адресу %s", address)

	// Запускаем сервер
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
