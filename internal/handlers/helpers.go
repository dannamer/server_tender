package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"net/http"
	"regexp"
	"strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
	"time"
)

func respondWithPanicError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{Reason: message})
	panic(message)
}

func validateName(w http.ResponseWriter, name string, fieldName string) {
	if name == "" {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("Название %s не может быть пустым", fieldName))
	}
	if len(name) > 100 {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("Название %s слишком длинное, максимум 100 символов", fieldName))
	}
}

func validateDescription(w http.ResponseWriter, description string, fieldName string) {
	if description == "" {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("Описание %s не может быть пустым", fieldName))
	}
	if len(description) > 500 {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("Описание %s слишком длинное, максимум 500 символов", fieldName))
	}
}

func validateID(w http.ResponseWriter, id string, idType string) {
	if id == "" {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("%s не может быть пустым", idType))

	}
	if len(id) > 100 {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("%s слишком длинное, максимум 100 символов", idType))
	}
	if _, err := uuid.Parse(id); err != nil {
		respondWithPanicError(w, http.StatusBadRequest, fmt.Sprintf("Некорректный формат %s", idType))
	}
}

func validateStatus(w http.ResponseWriter, status models.Status) {
	if status == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Статус не может быть пустым")
	}
	validStatuses := map[models.Status]bool{
		models.Created:   true,
		models.Published: true,
		models.Closed:    true,
	}
	if !validStatuses[status] {
		respondWithPanicError(w, http.StatusBadRequest, "Некорректный статус. Допустимые значения: Created, Published, Closed")
	}
}

func validateAuthorType(w http.ResponseWriter, authorType models.AuthorType) {
	if authorType == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Тип автора не может быть пустым")
	}
	validAuthorTypes := map[models.AuthorType]bool{
		models.AuthorTypeOrganization: true,
		models.AuthorTypeUser:         true,
	}
	if !validAuthorTypes[authorType] {
		respondWithPanicError(w, http.StatusBadRequest, "Некорректный тип автора. Допустимые значения: Organization, User")
	}
}

func validateServiceType(w http.ResponseWriter, serviceType models.ServiceType) {
	if serviceType == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Тип услуги не может быть пустым")
	}
	validServiceTypes := map[models.ServiceType]bool{
		models.Construction: true,
		models.Delivery:     true,
		models.Manufacture:  true,
	}
	if !validServiceTypes[serviceType] {
		respondWithPanicError(w, http.StatusBadRequest, "Некорректный тип услуги. Допустимые значения: Construction, Delivery, Manufacture")
	}
}

func validateDecision(w http.ResponseWriter, decision models.Сoordination) {
	if decision == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Решение не может быть пустым")
	}
	validDecisions := map[models.Сoordination]bool{
		models.Approved: true,
		models.Rejected: true,
	}
	if !validDecisions[decision] {
		respondWithPanicError(w, http.StatusBadRequest, "Некорректное решение. Допустимые значения: Approved, Rejected")
	}
}

func validateUsername(w http.ResponseWriter, username string) {
	if username == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Имя пользователя не может быть пустым")
	}
	if len(username) > 50 {
		respondWithPanicError(w, http.StatusBadRequest, "Имя пользователя слишком длинное, максимум 50 символов")
	}
	validUsernameRegex := `^[a-zA-Z0-9_-]+$`
	matched, err := regexp.MatchString(validUsernameRegex, username)
	if err != nil || !matched {
		respondWithPanicError(w, http.StatusBadRequest, "Имя пользователя содержит недопустимые символы. Разрешены только буквы, цифры, дефисы и символы подчеркивания")
	}
}

func validateFeedback(w http.ResponseWriter, feedback string) {
	if feedback == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Отзыв не может быть пустым")
	}
	if len(feedback) > 1000 {
		respondWithPanicError(w, http.StatusBadRequest, "Отзыв слишком длинный, максимум 1000 символов")
	}
}

func getAndValidateBidByID(w http.ResponseWriter, bidID string) *models.Bid {
	bid, err := database.GetBidByID(bidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusNotFound, fmt.Sprintf("Предложение с указанным ID %s не найдено", bidID))
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложения")
	}
	return bid
}

func getAndValidateUserByUsername(w http.ResponseWriter, username string) *models.User {
	user, err := database.GetUserByUsername(username)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusUnauthorized, fmt.Sprintf("Пользователь с именем '%s' не найден", username))
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении данных пользователя")
	}
	return user
}

func getAndValidateUserByID(w http.ResponseWriter, userID string) *models.User {
	user, err := database.GetUserByID(userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusUnauthorized, fmt.Sprintf("Пользователь с ID '%s' не найден", userID))
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении данных пользователя")
	}
	return user
}

func getAndValidateTenderByID(w http.ResponseWriter, tenderID string) *models.Tender {
	tender, err := database.GetTenderByID(tenderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusNotFound, fmt.Sprintf("Тендер с ID %s не найден", tenderID))
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
	}
	return tender
}

func getAndValidateTenderHistoryVersion(w http.ResponseWriter, tenderID string, version int) *models.TenderHistory {
	tenderHistoryVersion, err := database.GetTenderHistoryByVersion(tenderID, version)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusNotFound, "Версия тендера не найдена")
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении версии тендера")
	}
	return tenderHistoryVersion
}

func getAndValidateBidHistoryVersion(w http.ResponseWriter, bidID string, version int) *models.BidHistory {
	bidHistoryVersion, err := database.GetBidHistoryByVersion(bidID, version)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusNotFound, "Версия предложения не найдена")
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении версии предложения")
	}
	return bidHistoryVersion
}

func getAndValidateBidByTenderAndAuthorID(w http.ResponseWriter, tenderID, authorID string) *models.Bid {
	bid, err := database.GetBidByTenderAndAuthorID(tenderID, authorID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithPanicError(w, http.StatusNotFound, "Предложение не найдено")
		}
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложения")
	}
	return bid
}

func saveTenderHistory(w http.ResponseWriter, tender *models.Tender) {
	tenderHistory := &models.TenderHistory{
		TenderID:    tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
	}
	if err := database.SaveTenderHistory(tenderHistory); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при сохранении истории тендера")
	}
}

func saveBidHistory(w http.ResponseWriter, bid *models.Bid) {
	bidHistory := &models.BidHistory{
		BidID:       bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
		Version:     bid.Version,
	}
	if err := database.SaveBidHistory(bidHistory); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при сохранении истории тендера")
	}
}

func createTender(tender *models.TenderRequest) *models.Tender {
	user, _ := database.GetUserByUsername(tender.CreatorUsername)
	return &models.Tender{
		Name:              tender.Name,
		Description:       tender.Description,
		ServiceType:       tender.ServiceType,
		Status:            models.Created,
		OrganizationID:    tender.OrganizationID,
		CreatorUsernameID: user.ID,
		Version:           1,
	}
}

func createBid(bid *models.BidRequest) *models.Bid {
	return &models.Bid{
		Name:         bid.Name,
		Description:  bid.Description,
		Status:       models.Created,
		TenderID:     bid.TenderID,
		AuthorType:   bid.AuthorType,
		AuthorID:     bid.AuthorID,
		Version:      1,
		Сoordination: models.Expectation,
	}
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

func createBidResponse(bid *models.Bid) *models.BidResponse {
	return &models.BidResponse{
		ID:          bid.ID,
		Name:        bid.Name,
		Description: bid.Description,
		Status:      bid.Status,
		TenderID:    bid.ID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
		Version:     bid.Version,
		CreatedAt:   bid.CreatedAt.Format(time.RFC3339),
	}
}

func updateTenderFields(w http.ResponseWriter, tenderEditRequest *models.TenderEditRequest, tender *models.Tender) {
	if tenderEditRequest.Name != "" {
		validateName(w, tenderEditRequest.Name, "тендера")
		tender.Name = tenderEditRequest.Name
	}
	if tenderEditRequest.Description != "" {
		validateDescription(w, tenderEditRequest.Description, "тендера")
		tender.Description = tenderEditRequest.Description
	}
	if tenderEditRequest.ServiceType != "" {
		validateServiceType(w, tenderEditRequest.ServiceType)
		tender.ServiceType = tenderEditRequest.ServiceType
	}
}

func updateBidFields(w http.ResponseWriter, bidEditRequest *models.BidEditRequest, bid *models.Bid) {
	if bidEditRequest.Name != "" {
		validateName(w, bidEditRequest.Name, "предложения")
		bid.Name = bidEditRequest.Name
	}
	if bidEditRequest.Description != "" {
		validateDescription(w, bidEditRequest.Description, "предложения")
	}
}

func validateLimitAndOffset(w http.ResponseWriter, limitParam, offsetParam string) (int, int) {
	limit, offset := 5, 0
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			respondWithPanicError(w, http.StatusBadRequest, "Некорректный параметр лимита")
		}
	}
	if offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else {
			respondWithPanicError(w, http.StatusBadRequest, "Некорректный параметр смещения")
		}
	}
	return limit, offset
}

func validateVersion(w http.ResponseWriter, versionParam string, currentVersion int) int {
	version, err := strconv.Atoi(versionParam)
	if err != nil || version < 1 || currentVersion <= version {
		respondWithPanicError(w, http.StatusBadRequest, "Некорректная версия")
	}
	return version
}
