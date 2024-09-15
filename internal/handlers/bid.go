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
	
	tender := getAndValidateTenderByID(w, bidRequest.TenderID)
	user := getAndValidateUserByUsername(w, bidRequest.AuthorID)

	if bidRequest.AuthorType == models.AuthorTypeOrganization {
		if !database.HasUserOrganization(user.ID) {
			respondWithPanicError(w, http.StatusForbidden, "Пользователь не связан с организацией")
		}
	}
	

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
	
	user := getAndValidateUserByUsername(w, username)

	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)

	bids, err := database.GetBidsByUserID(user.ID, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)
}

func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	validateID(w, tenderID, "ID тендера")
	validateUsername(w, username)

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")
	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)
	
	user := getAndValidateUserByUsername(w, username)
	tender := getAndValidateTenderByID(w, tenderID)
	
	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
	}
	bids, err := database.GetBidsByTenderID(tenderID, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bids)
}

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

	w.Header().Set("Content-Type", "application/json")
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

	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
		}
	} else if user.ID != bid.AuthorID {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
	}

	var editBidHandler models.BidEditRequest
	if err := json.NewDecoder(r.Body).Decode(&editBidHandler); err != nil {
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

func RollbackBidHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	versionParam := chi.URLParam(r, "version")
	username := r.URL.Query().Get("username")

	validateID(w, bidID, "ID тендера")
	validateUsername(w, username)

	bid := getAndValidateBidByID(w, bidID)
	user := getAndValidateUserByUsername(w, username)

	version := validateVersion(w, versionParam, bid.Version)

	if bid.AuthorType == models.AuthorTypeOrganization {
		if !database.CheckUserOrganizationResponsibility(user.ID, bid.AuthorID) {
			respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации предложения")
		}
	} else if user.ID != bid.AuthorID {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет доступа к предложению")
	}

	bidHistory := getAndValidateBidHistoryVersion(w, bid.ID, version)

	saveBidHistory(w, bid)

	bid.Name = bidHistory.Name
	bid.Description = bidHistory.Description
	bid.Version++

	if err := database.UpdateBid(bid); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}

func SubmitBidFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	bidID := chi.URLParam(r, "bidId")
	bidFeedback := r.URL.Query().Get("bidFeedback")
	username := r.URL.Query().Get("username")

	validateID(w, bidID, "ID предложения")
	validateUsername(w, username)
	validateFeedback(w, bidFeedback)

	// Получаем пользователя по username
	user := getAndValidateUserByUsername(w, username)
	bid := getAndValidateBidByID(w, bidID)
	teder := getAndValidateTenderByID(w, bid.TenderID)

	if !database.CheckUserOrganizationResponsibility(user.ID, teder.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации тендера")
	}

	if bid.Status == models.Closed {
		respondWithPanicError(w, http.StatusForbidden, "Отзыв может быть отправлен только на опубликованное или закрытое предложение")
	}

	feedback := &models.Feedback{
		UserID:      user.ID,
		BidID:       bid.ID,
		BidFeedback: bidFeedback,
	}

	if err := database.SaveFeedback(feedback); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при сохранении отзыва")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createBidResponse(bid))
}

func GetBidReviewsHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	authorUsername := r.URL.Query().Get("authorUsername")
	requesterUsername := r.URL.Query().Get("requesterUsername")

	validateID(w, tenderID, "ID тендера")
	validateUsername(w, authorUsername)
	validateUsername(w, requesterUsername)

	authorUser := getAndValidateUserByUsername(w, authorUsername)
	requesterUser := getAndValidateUserByUsername(w, requesterUsername)

	tender := getAndValidateTenderByID(w, tenderID)
	bid := getAndValidateBidByTenderAndAuthorID(w, tender.ID, authorUser.ID)

	if !database.CheckUserOrganizationResponsibility(requesterUser.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для просмотра отзывов на предложения")
	}

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)

	feedbackResponse, err := database.GetBidReviews(bid.ID, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении отзывов")
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(feedbackResponse)
}
