package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"tender-service/internal/database"
	"tender-service/internal/models"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderRequest := &models.TenderRequest{}
	if err := json.NewDecoder(r.Body).Decode(tenderRequest); err != nil {
		respondWithPanicError(w, http.StatusBadRequest, "Неверный формат запроса")
	}

	validateName(w, tenderRequest.Name, "тендера")
	validateDescription(w, tenderRequest.Description, "тендера")
	validateServiceType(w, models.ServiceType(tenderRequest.ServiceType))
	validateID(w, tenderRequest.OrganizationID, "ID организации")
	validateUsername(w, tenderRequest.CreatorUsername)

	user := getAndValidateUserByUsername(w, tenderRequest.CreatorUsername)

	if !database.CheckUserOrganizationResponsibility(user.ID, tenderRequest.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Пользователь не имеет прав для создания тендеров от имени этой организации")
	}

	tender := createTender(tenderRequest)

	if err := database.SaveTender(tender); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при сохранении тендера")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}

func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")
	serviceTypes := r.URL.Query()["service_type"]

	for _, serviceType := range serviceTypes {
		validateServiceType(w, models.ServiceType(serviceType))
	}

	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)

	tenders, err := database.GetTendersResponse(serviceTypes, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении тендеров")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}

func GetUserTendersHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	username := r.URL.Query().Get("username")
	user := getAndValidateUserByUsername(w, username)

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit, offset := validateLimitAndOffset(w, limitParam, offsetParam)

	tenders, err := database.GetTendersByUsername(user.ID, limit, offset)
	if err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при получении тендеров пользователя")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}

func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	validateID(w, tenderID, "ID тендера")
	if username != "" {
		getAndValidateUserByUsername(w, username)
	}

	tender := getAndValidateTenderByID(w, tenderID)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tender.Status)
}

func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	newStatus := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")

	validateID(w, tenderID, "ID тендера")
	validateStatus(w, models.Status(newStatus))
	validateUsername(w, username)

	user := getAndValidateUserByUsername(w, username)

	tender := getAndValidateTenderByID(w, tenderID)

	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Пользователь не имеет отвественности за организацию текущего тендера для изменения статуса")
	}

	tender.Status = models.Status(newStatus)

	if err := database.UpdateTender(tender); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса тендера")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	validateID(w, tenderID, "ID тендера")
	validateUsername(w, username)

	tender := getAndValidateTenderByID(w, tenderID)
	user := getAndValidateUserByUsername(w, username)

	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для редактирования тендера")
	}

	tenderEditRequest := &models.TenderEditRequest{}
	err := json.NewDecoder(r.Body).Decode(tenderEditRequest)
	if err != nil {
		respondWithPanicError(w, http.StatusBadRequest, "Неверный формат запроса")
	}

	if tenderEditRequest.Name == "" && tenderEditRequest.Description == "" && tenderEditRequest.ServiceType == "" {
		respondWithPanicError(w, http.StatusBadRequest, "Отправлен пустой запрос")
	}

	copyTender := *tender

	updateTenderFields(w, tenderEditRequest, tender)
	saveTenderHistory(w, &copyTender)

	tender.Version++

	if err := database.UpdateTender(tender); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}

func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		recover()
	}()

	tenderID := chi.URLParam(r, "tenderId")
	versionParam := chi.URLParam(r, "version")
	username := r.URL.Query().Get("username")

	validateID(w, tenderID, "ID тендера")
	validateUsername(w, username)

	tender := getAndValidateTenderByID(w, tenderID)
	user := getAndValidateUserByUsername(w, username)

	version := validateVersion(w, versionParam, tender.Version)

	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithPanicError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
	}

	tenderHistoryVersion := getAndValidateTenderHistoryVersion(w, tender.ID, version)

	saveTenderHistory(w, tender)

	tender.Name = tenderHistoryVersion.Name
	tender.Description = tenderHistoryVersion.Description
	tender.ServiceType = tenderHistoryVersion.ServiceType
	tender.Version++

	if err := database.UpdateTender(tender); err != nil {
		respondWithPanicError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}