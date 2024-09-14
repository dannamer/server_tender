package handlers

import (
	"encoding/json"
	// "fmt"
	"github.com/go-chi/chi/v5"
	// "github.com/jackc/pgx/v4"
	"net/http"
	// "strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
	// "time"
)

// исправить
func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidRequest := &models.BidRequest{}
	if err := json.NewDecoder(r.Body).Decode(bidRequest); err != nil {
		respondWithPanicError(w, http.StatusBadRequest, "Неверный формат данных")
	}

	validateAuthorType(w, bidRequest.AuthorType)
	validateID(w, bidRequest.AuthorID, "ID автора")
	validateID(w, bidRequest.TenderID, "ID тендера")
	validateName(w, bidRequest.Name, "предложения")
	validateDescription(w, bidRequest.Description, "предложения")

	if bidRequest.AuthorType == models.AuthorTypeOrganization {
		getAndValidateOrganizationByID(w, bidRequest.AuthorID)
	} else {
		getAndValidateUserByID(w, bidRequest.AuthorID)
	}

	tender := getAndValidateTenderByID(w, bidRequest.TenderID)

	if tender.Status != models.Published {
		respondWithPanicError(w, http.StatusForbidden, "Тендер ещё не опубликован или закрыт")
	}

	bid := createBid(bidRequest)

	if err := database.SaveBid(bid); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при создании предложения")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}

func GetUserBidsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	username := r.URL.Query().Get("username")
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	getAndValidateUserByUsername(w, username)

	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)

	bids, err := database.GetBidsByUsername(username, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)
}

//исправить
// func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
// 	// Проверка метода запроса
// 	if r.Method != http.MethodGet {
// 		respondWithPanicError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
// 		return
// 	}

// 	// Получаем параметры из URL
// 	tenderID := chi.URLParam(r, "tenderId")
// 	username := r.URL.Query().Get("username")

// 	// Проверяем, заданы ли обязательные параметры
// 	if tenderID == "" || username == "" {
// 		respondWithPanicError(w, http.StatusBadRequest, "Отсутствуют обязательные параметры: tenderId или username")
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
// 			respondWithPanicError(w, http.StatusBadRequest, "Некорректное значение limit")
// 			return
// 		}
// 	}
// 	if offsetParam != "" {
// 		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
// 			offset = parsedOffset
// 		} else {
// 			respondWithPanicError(w, http.StatusBadRequest, "Некорректное значение offset")
// 			return
// 		}
// 	}

// 	userExists, err := database.CheckUserExists(username)
// 	if err != nil {
// 		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка проверки пользователя")
// 		return
// 	}

// 	if !userExists {
// 		respondWithPanicError(w, http.StatusUnauthorized, "Пользователь не существует")
// 		return
// 	}

// 	// Проверяем, существует ли тендер
// 	tender, err := database.GetTenderByID(tenderID)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			respondWithPanicError(w, http.StatusNotFound, "Тендер не найден")
// 			return
// 		}
// 		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
// 		return
// 	}

// 	// Проверяем права пользователя
// 	if !database.CheckUserOrganizationResponsibility(username, tender.OrganizationID) {
// 		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
// 		return
// 	}

// 	// Получаем список предложений для указанного тендера
// 	bids, err := database.GetBidsByTenderID(tenderID, limit, offset)
// 	if err != nil {
// 		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
// 		return
// 	}

// 	// Если предложений нет, возвращаем 404
// 	if len(bids) == 0 {
// 		respondWithPanicError(w, http.StatusNotFound, "Предложения не найдены")
// 		return
// 	}

// 	// Возвращаем список предложений
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(bids)
// }

func SubmitBidDecisionHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	decision := models.Сoordination(r.URL.Query().Get("decision"))
	username := r.URL.Query().Get("username")

	validateID(w, bidID, "ID предложения")
	validateDecision(w, decision)
	validateUsername(w, username)

	user := getAndValidateUserByUsername(w, username)
	bid := getAndValidateBidByID(w, bidID)
	tender := getAndValidateTenderByID(w, bid.TenderID)

	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации стенда")
	}
	if bid.Status != models.Published {
		respondWithPanicError(w, http.StatusForbidden, "Предложение не может быть обработано, так как его статус находится в состоянии 'Создание' или 'Закрыто'. Решение может быть отправлено только для предложений в статусе 'Публичное'.")
	}
	if tender.Status != models.Published {
		respondWithPanicError(w, http.StatusForbidden, "Тендер не может быть обработан, так как он находится в статусе 'Создание' или 'Закрыт'. Для отправление решении тендер должен находиться в статусе 'Публичный'.")
	}

	userDecision := &models.UserDecision{
		UserID:   user.ID,
		BidID:    bid.ID,
		Decision: decision,
	}

	if database.CheckUserDecisionExists(userDecision) {
		respondWithPanicError(w, http.StatusForbidden, "пользователь раннее давал свое решение")
	}

	if err := database.SaveUserDecision(userDecision); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при сохранении решения")
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
			respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
		}
		for _, Bid := range bids {
			Bid.Status = models.Closed
			Bid.Сoordination = models.RejectedByConflict
			if err := database.UpdateBid(&Bid); err != nil {
				respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса предложения")
			}
		}
	}

	if err := database.UpdateBid(bid); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса предложения")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}

func GetBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	username := r.URL.Query().Get("username")

	validateID(w, bidID, "ID предложения")
	validateUsername(w, username)

	user := getAndValidateUserByUsername(w, username)
	bid := getAndValidateBidByID(w, bidID)

	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
		}
	} else if user.ID != bid.AuthorID {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bid.Status)
}

func UpdateBidStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	newStatus := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	validateID(w, bidID, "ID предложения")
	validateStatus(w, models.Status(newStatus))
	validateUsername(w, username)

	user := getAndValidateUserByUsername(w, username)
	bid := getAndValidateBidByID(w, bidID)

	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
		}
	} else if user.ID != bid.AuthorID {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
	}

	bid.Status = models.Status(newStatus)

	if err := database.UpdateBid(bid); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса тендера")
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}

func EditBidHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	username := r.URL.Query().Get("username")

	validateID(w, bidID, "ID предложения") 
	validateUsername(w, username)

	bid := getAndValidateBidByID(w, bidID)
	user := getAndValidateUserByUsername(w, username)

	if bid.AuthorType == models.AuthorTypeUser && bid.AuthorID != user.ID {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для редактирования тендера")
	} else if bid.AuthorType == models.AuthorTypeOrganization && !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для редактирования тендера")
	}

	var editBidHandler models.BidEditRequest
	err := json.NewDecoder(r.Body).Decode(&editBidHandler)
	if err != nil {
		respondWithPanicError(w, http.StatusBadRequest, "Неверный формат запроса")
	}
	if editBidHandler.Name == "" && editBidHandler.Description == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Отправлен пустой запрос")
	}
	copybid := *bid
	updateBidFields(w, &editBidHandler, bid)
	saveBidHistory(w, &copybid)
	bid.Version++
	if err := database.UpdateBid(bid); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}
