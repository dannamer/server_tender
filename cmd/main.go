package main

import (
	"log"
	"net/http"
	// "tender-service/config"
	"tender-service/internal/handlers"
	"tender-service/internal/database"
	"github.com/go-chi/chi/v5"
)

func main() {
	_, err := database.ConnectPostgres()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	r := chi.NewRouter()
	// Привязываем обработчики для API
	r.Get("/api/ping", handlers.PingHandler)                    // Проверка доступности сервера
	r.Post("/api/tenders/new", handlers.CreateTenderHandler)       // Создание нового тендера
	r.Get("/api/tenders/{tenderId}/status", handlers.GetTenderStatusHandler) // Получение статуса тендера
	r.Put("/api/tenders/{tenderId}/status", handlers.UpdateTenderStatusHandler)
	r.Get("/api/tenders", handlers.GetTendersHandler)            // Получение списка тендеров
	r.Get("/api/tenders/my", handlers.GetUserTendersHandler)     // Получение тендеров пользователя
	r.Patch("/api/tenders/{tenderId}/edit", handlers.EditTenderHandler) // Редактирование тендера
	r.Put("/api/tenders/{tenderId}/rollback/{version}", handlers.RollbackTenderHandler) // Откат тендера к версии
	r.Post("/api/bids/new", handlers.CreateBidHandler)            // Создание нового предложения
	r.Get("/api/bids/my", handlers.GetUserBidsHandler)           // Получение предложений пользователя
	// r.Get("/api/bids/{tenderId}/list", handlers.GetBidsForTenderHandler) // Получение предложений для тендера
	r.Get("/api/bids/{bidId}/status", handlers.GetBidStatusHandler)  // Получение статуса предложения
	r.Put("/api/bids/{bidId}/status", handlers.UpdateBidStatusHandler)
	r.Put("/api/bids/{bidId}/submit_decision", handlers.SubmitBidDecisionHandler) // Отправка решения по предложению
	r.Patch("/api/bids/{bidId}/edit", handlers.EditBidHandler)
	r.Put("/api/bids/{bidId}/rollback/{version}", handlers.RollbackBidHandler)
	r.Put("/api/bids/{bidId}/feedback", handlers.GetBidReviewsHandler)
	r.Get("/api/bids/{tenderId}/reviews", handlers.SubmitBidDecisionHandler)
	address := "0.0.0.0:8080"
	log.Printf("Сервер запущен по адресу %s", address)

	// Запускаем сервер
	if err := http.ListenAndServe(address, r); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}
