package handlers

import (
	"encoding/json"
	"net/http"
	"tender-service/internal/models"
	"tender-service/internal/database"
	"time"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		http.Error(w, "Unable to write response", http.StatusInternalServerError)
	}
}

// CreateTenderHandler обрабатывает создание тендера
func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	var tender models.Tender

	// Декодируем тело запроса в структуру Tender
	err := json.NewDecoder(r.Body).Decode(&tender)
	if err != nil {
		http.Error(w, "Неверные данные в запросе", http.StatusBadRequest)
		return
	}

	// Проверяем, переданы ли необходимые поля
	if tender.Name == "" || tender.OrganizationID == 0 || tender.CreatorID == 0 {
		http.Error(w, "Отсутствуют обязательные поля", http.StatusBadRequest)
		return
	}

	// Проверяем, является ли пользователь ответственным за организацию
	isResponsible := database.CheckUserOrganizationResponsibility(tender.CreatorID, tender.OrganizationID)
	if !isResponsible {
		http.Error(w, "Пользователь не имеет прав для создания тендеров от имени этой организации", http.StatusForbidden)
		return
	}

	// Устанавливаем значения по умолчанию
	tender.CreatedAt = time.Now()
	tender.UpdatedAt = time.Now()
	tender.Version = 1

	// Сохраняем тендер в базу данных
	err = database.SaveTender(&tender)
	if err != nil {
		http.Error(w, "Ошибка при сохранении тендера", http.StatusInternalServerError)
		return
	}

	// Возвращаем созданный тендер в ответ
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tender)
}
