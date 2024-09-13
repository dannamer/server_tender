package handlers

import (
	"encoding/json"
	// "fmt"
	"github.com/go-chi/chi/v5"
	// "github.com/jackc/pgx/v4"
	"net/http"
	"strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
	"time"
)

// исправить
func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	// Декодируем тело запроса
	///
	var bidRequest models.BidRequest
	err := json.NewDecoder(r.Body).Decode(&bidRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}
	if validateAuthorType(w, bidRequest.AuthorType) || validateID(w, bidRequest.AuthorID, "ID автора") || validateID(w, bidRequest.TenderID, "ID тендера") || validateName(w, bidRequest.Name, "предложения") || validateDescription(w, bidRequest.Description, "предложения") {
		return
	}

	if bidRequest.AuthorType == models.AuthorTypeOrganization {
		_, res := getAndValidateOrganizationByID(w, bidRequest.AuthorID)
		if res {
			return
		}
	} else if _, res := getAndValidateUserByID(w, bidRequest.AuthorID); res {
		return
	}
	tender, res := getAndValidateTenderByID(w, bidRequest.TenderID)
	if res {
		return
	}

	if tender.Status != models.Published {
		respondWithError(w, http.StatusForbidden, "Тендер ещё не опубликован или закрыт")
		return
	}

	bid := models.Bid{
		Name:         bidRequest.Name,
		Description:  bidRequest.Description,
		Status:       models.Created,
		TenderID:     bidRequest.TenderID,
		AuthorType:   bidRequest.AuthorType,
		AuthorID:     bidRequest.AuthorID,
		Version:      1,
		Сoordination: models.Expectation,
	}

	err = database.SaveBid(&bid)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при создании предложения")
		return
	}
	bidResponse := models.BidResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
	}
	// Возвращаем успешный ответ с информацией о созданном предложении
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponse)
}

// исправить
// GetUserBidsHandler обработчик для получения списка предложений текущего пользователя
func GetUserBidsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры username, limit и offset
	username := r.URL.Query().Get("username")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Проверяем наличие имени пользователя
	if username == "" {
		respondWithError(w, http.StatusUnauthorized, "Необходимо указать имя пользователя")
		return
	}

	// Проверяем, существует ли пользователь
	exists, err := database.EmployeeExists(username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при проверке существования пользователя")
		return
	}
	if !exists {
		respondWithError(w, http.StatusUnauthorized, "Пользователь не найден")
		return
	}

	// Устанавливаем лимит и смещение по умолчанию
	limit := 5
	offset := 0

	// Обрабатываем параметры пагинации
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректное значение лимита")
			return
		}
	}

	if offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректное значение смещения")
			return
		}
	}

	// Получаем список предложений пользователя
	bids, err := database.GetBidsByUsername(username, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
		return
	}

	// Возвращаем список предложений
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)
}

//исправить
// func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
// 	// Проверка метода запроса
// 	if r.Method != http.MethodGet {
// 		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
// 		return
// 	}

// 	// Получаем параметры из URL
// 	tenderID := chi.URLParam(r, "tenderId")
// 	username := r.URL.Query().Get("username")

// 	// Проверяем, заданы ли обязательные параметры
// 	if tenderID == "" || username == "" {
// 		respondWithError(w, http.StatusBadRequest, "Отсутствуют обязательные параметры: tenderId или username")
// 		return
// 	}

// 	// Получаем параметры пагинации
// 	limitParam := r.URL.Query().Get("limit")
// 	offsetParam := r.URL.Query().Get("offset")
// 	limit := 5  // Значение по умолчанию
// 	offset := 0 // Значение по умолчанию

// 	// Парсим limit и offset
// 	if limitParam != "" {
// 		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
// 			limit = parsedLimit
// 		} else {
// 			respondWithError(w, http.StatusBadRequest, "Некорректное значение limit")
// 			return
// 		}
// 	}
// 	if offsetParam != "" {
// 		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
// 			offset = parsedOffset
// 		} else {
// 			respondWithError(w, http.StatusBadRequest, "Некорректное значение offset")
// 			return
// 		}
// 	}

// 	userExists, err := database.CheckUserExists(username)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Ошибка проверки пользователя")
// 		return
// 	}

// 	if !userExists {
// 		respondWithError(w, http.StatusUnauthorized, "Пользователь не существует")
// 		return
// 	}

// 	// Проверяем, существует ли тендер
// 	tender, err := database.GetTenderByID(tenderID)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			respondWithError(w, http.StatusNotFound, "Тендер не найден")
// 			return
// 		}
// 		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
// 		return
// 	}

// 	// Проверяем права пользователя
// 	if !database.CheckUserOrganizationResponsibility(username, tender.OrganizationID) {
// 		respondWithError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
// 		return
// 	}

// 	// Получаем список предложений для указанного тендера
// 	bids, err := database.GetBidsByTenderID(tenderID, limit, offset)
// 	if err != nil {
// 		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
// 		return
// 	}

// 	// Если предложений нет, возвращаем 404
// 	if len(bids) == 0 {
// 		respondWithError(w, http.StatusNotFound, "Предложения не найдены")
// 		return
// 	}

// 	// Возвращаем список предложений
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(bids)
// }

// готово
func SubmitBidDecisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}
	bidID := chi.URLParam(r, "bidId")
	decision := models.Сoordination(r.URL.Query().Get("decision"))
	username := r.URL.Query().Get("username")
	if validateID(w, bidID, "ID предложения") || validateDecision(w, decision) || validateUsername(w, username) {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	bid, res := getAndValidateBidByID(w, bidID)
	if res {
		return
	}
	tender, res := getAndValidateTenderByID(w, bid.TenderID)
	if res {
		return
	}
	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации стенда")
	}
	if bid.Status != models.Published {
		respondWithError(w, http.StatusForbidden, "Предложение не может быть обработано, так как его статус находится в состоянии 'Создание' или 'Закрыто'. Решение может быть отправлено только для предложений в статусе 'Публичное'.")
		return
	}
	if tender.Status != models.Published {
		respondWithError(w, http.StatusForbidden, "Тендер не может быть обработан, так как он находится в статусе 'Создание' или 'Закрыт'. Для отправление решении тендер должен находиться в статусе 'Публичный'.")
		return
	}
	// записываем решение
	userDecision := &models.UserDecision{
		UserID:   user.ID,
		BidID:    bid.ID,
		Decision: decision,
	}
	if database.CheckUserDecisionExists(userDecision) {
		respondWithError(w, http.StatusForbidden, "пользователь раннее давал свое решение")
		return
	}
	if err := database.SaveUserDecision(userDecision); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при сохранении решения")
		return
	}
	bid.UserDecision = append(bid.UserDecision, *userDecision)
	if decision == models.Rejected {
		bid.Сoordination = models.Rejected
		bid.Status = models.Closed
	} else if len(bid.UserDecision) >= 3 {
		bid.Сoordination = models.Approved
		bid.Status = models.Closed
		bids, err := database.GetBidsByTenderIDWithExpectation(tender.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
			return
		}
		for _, bida := range bids {
			bida.Status = models.Closed
			bida.Сoordination = models.RejectedByConflict
			if err := database.UpdateBid(&bida); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса предложения")
				return
			}
		}
	}
	if err := database.UpdateBid(bid); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса предложения")
		return
	}

	bidResponse := &models.BidResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponse)
}

// готово
func GetBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	bidID := chi.URLParam(r, "bidId")
	username := r.URL.Query().Get("username")

	if validateID(w, bidID, "ID предложения") || validateUsername(w, username) {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	bid, res := getAndValidateBidByID(w, bidID)
	if res {
		return
	}
	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
			return
		}
	} else if user.ID == bid.AuthorID {
		respondWithError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid.Status)
}

// готово
func UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	bidID := chi.URLParam(r, "bidId")
	newStatus := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	if validateID(w, bidID, "ID предложения") || validateStatus(w, models.Status(newStatus)) || validateUsername(w, username) {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	bid, res := getAndValidateBidByID(w, bidID)
	if res {
		return
	}
	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
			return
		}
	} else if user.ID == bid.AuthorID {
		respondWithError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
		return
	}
	bid.Status = models.Status(newStatus)
	if err := database.UpdateBid(bid); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса тендера")
		return
	}
	bidResponse := &models.BidResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt.Format(time.RFC3339),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponse)
}
