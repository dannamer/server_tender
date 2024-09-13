package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"net/http"
	"regexp"
	"time"
	"tender-service/internal/database"
	"tender-service/internal/models"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{Reason: message})
}

func validateName(w http.ResponseWriter, name string, fieldName string) bool {
	if name == "" {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Название %s не может быть пустым", fieldName))
		return true
	}
	if len(name) > 100 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Название %s слишком длинное, максимум 100 символов", fieldName))
		return true
	}
	return false
}

func validateDescription(w http.ResponseWriter, description string, fieldName string) bool {
	if description == "" {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Описание %s не может быть пустым", fieldName))
		return true
	}
	if len(description) > 500 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Описание %s слишком длинное, максимум 500 символов", fieldName))
		return true
	}
	return false
}

func validateID(w http.ResponseWriter, id string, idType string) bool {
	if id == "" {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%s не может быть пустым", idType))
		return true
	}
	if len(id) > 100 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%s слишком длинное, максимум 100 символов", idType))
		return true
	}
	if _, err := uuid.Parse(id); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Некорректный формат %s", idType))
		return true
	}
	return false
}

func validateStatus(w http.ResponseWriter, status models.Status) bool {
	if status == "" {
		respondWithError(w, http.StatusBadRequest, "Статус не может быть пустым")
		return true
	}
	validStatuses := map[models.Status]bool{
		models.Created:   true,
		models.Published: true,
		models.Closed:  true,
	}
	if !validStatuses[status] {
		respondWithError(w, http.StatusBadRequest, "Некорректный статус. Допустимые значения: Created, Published, Closed")
		return true
	}
	return false
}

func validateServiceType(w http.ResponseWriter, serviceType models.ServiceType) bool {
	if serviceType == "" {
		respondWithError(w, http.StatusBadRequest, "Тип услуги не может быть пустым")
		return true
	}
	validServiceTypes := map[models.ServiceType]bool{
		models.Construction: true,
		models.Delivery:     true,
		models.Manufacture:  true,
	}
	if !validServiceTypes[serviceType] {
		respondWithError(w, http.StatusBadRequest, "Некорректный тип услуги. Допустимые значения: Construction, Delivery, Manufacture")
		return true
	}
	return false
}

func validateDecision(w http.ResponseWriter, decision models.Сoordination) bool {
	if decision == "" {
		respondWithError(w, http.StatusBadRequest, "Решение не может быть пустым")
		return true
	}
	validDecisions := map[models.Сoordination]bool{
		models.Approved: true,
		models.Rejected: true,
	}
	if !validDecisions[decision] {
		respondWithError(w, http.StatusBadRequest, "Некорректное решение. Допустимые значения: Approved, Rejected")
		return true
	}
	return false
}

func validateUsername(w http.ResponseWriter, username string) bool {
	if username == "" {
		respondWithError(w, http.StatusBadRequest, "Имя пользователя не может быть пустым")
		return true
	}
	if len(username) > 50 {
		respondWithError(w, http.StatusBadRequest, "Имя пользователя слишком длинное, максимум 50 символов")
		return true
	}
	validUsernameRegex := `^[a-zA-Z0-9_-]+$`
	matched, err := regexp.MatchString(validUsernameRegex, username)
	if err != nil || !matched {
		respondWithError(w, http.StatusBadRequest, "Имя пользователя содержит недопустимые символы. Разрешены только буквы, цифры, дефисы и символы подчеркивания")
		return true
	}
	return false
}

func getAndValidateBidByID(w http.ResponseWriter, bidID string) (*models.Bid, bool) {
	bid, err := database.GetBidByID(bidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Если предложение не найдено
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Предложение с указанным ID %s не найдено", bidID))
			return nil, true
		}
		// Если произошла другая ошибка базы данных
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении предложения")
		return nil, true
	}
	return bid, false
}

func getAndValidateUserByUsername(w http.ResponseWriter, username string) (*models.User, bool) {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Пользователь с именем '%s' не найден", username))
			return nil, true
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении данных пользователя")
		return nil, true
	}
	return user, false
}

func getAndValidateTenderByID(w http.ResponseWriter, tenderID string) (*models.Tender, bool) {
	tender, err := database.GetTenderByID(tenderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Тендер с ID %s не найден", tenderID))
			return nil, true
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
		return nil, true
	}
	return tender, false
}

func saveTenderHistory(w http.ResponseWriter, tender *models.Tender) bool {
	tenderHistory := &models.TenderHistory{
		TenderID:    tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
	}
	if err := database.SaveTenderHistory(tenderHistory); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при сохранении истории тендера")
		return true
	}
	return false
}

func createTenderResponse(tender *models.Tender) *models.TenderResponse {
	return &models.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt.Format(time.RFC3339),
	}
}

func updateTenderFields(w http.ResponseWriter, tenderEditRequest *models.TenderEditRequest, tender *models.Tender) bool {
	if tenderEditRequest.Name != "" {
		if validateName(w, tenderEditRequest.Name, "тендера") {
			return true
		}
		tender.Name = tenderEditRequest.Name
	}

	if tenderEditRequest.Description != "" {
		if validateDescription(w, tenderEditRequest.Description, "тендера") {
			return true
		}
		tender.Description = tenderEditRequest.Description
	}

	if tenderEditRequest.ServiceType != "" {
		if validateServiceType(w, tenderEditRequest.ServiceType) {
			return true
		}
		tender.ServiceType = tenderEditRequest.ServiceType
	}

	return false
}