package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	"net/http"
	"strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		http.Error(w, "Невозможно написать ответ", http.StatusInternalServerError)
	}
}

// готово
func CreateTenderHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	var tenderRequest models.TenderRequest

	// Декодируем тело запроса
	err := json.NewDecoder(r.Body).Decode(&tenderRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}
	if validateName(w, tenderRequest.Name, "тендера") || validateDescription(w, tenderRequest.Description, "тендера") || validateServiceType(w, models.ServiceType(tenderRequest.ServiceType)) || validateID(w, tenderRequest.OrganizationID, "ID организации") || validateUsername(w, tenderRequest.CreatorUsername) {
		return
	}

	// Проверяем права пользователя
	user, res := getAndValidateUserByUsername(w, tenderRequest.CreatorUsername)
	if res {
		return
	}
	if !database.CheckUserOrganizationResponsibility(user.ID, tenderRequest.OrganizationID) {
		respondWithError(w, http.StatusForbidden, "Пользователь не имеет прав для создания тендеров от имени этой организации")
		return
	}

	// Создаем объект тендера
	tender := models.Tender{
		Name:              tenderRequest.Name,
		Description:       tenderRequest.Description,
		ServiceType:       tenderRequest.ServiceType,
		Status:            models.Created,
		OrganizationID:    tenderRequest.OrganizationID,
		CreatorUsernameID: user.ID,
		Version:           1,
	}

	// Сохраняем тендер в базу данных
	if err = database.SaveTender(&tender); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при сохранении тендера")
		return
	}
	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(&tender))
}
//вроде как готово
func GetTendersHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры пагинации и тип услуги из query-параметров
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")
	serviceTypes := r.URL.Query()["service_type"] // Массив значений параметра service_type
	for _, serviceType := range serviceTypes {
		if validateServiceType(w, models.ServiceType(serviceType)) {
			return
		}
	}
	// Значения по умолчанию для пагинации
	limit := 5
	offset := 0

	// Обрабатываем лимит и смещение (если заданы)
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректный параметр лимита")
			return
		}
	}
	if offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректный параметр смещения")
			return
		}
	}

	// Получаем список тендеров из базы данных через функцию
	tenders, err := database.GetTendersResponse(serviceTypes, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендеров")
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}
//вроде как готово
func GetUserTendersHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем username из query-параметров
	username := r.URL.Query().Get("username")
	if username == "" {
		respondWithError(w, http.StatusBadRequest, "Отсутствует параметр username")
		return
	}

	// Получаем параметры пагинации
	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	// Значения по умолчанию для пагинации
	limit := 5
	offset := 0

	// Обрабатываем лимит и смещение (если заданы)
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректный параметр лимита")
			return
		}
	}
	if offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		} else {
			respondWithError(w, http.StatusBadRequest, "Некорректный параметр смещения")
			return
		}
	}

	// Получаем список тендеров пользователя из базы данных
	tenders, err := database.GetTendersByUsername(username, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендеров пользователя")
		return
	}

	// Возвращаем список тендеров в формате JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}

// готово
func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из URL (tenderId и username)
	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	if validateID(w, tenderID, "ID тендера") || validateUsername(w, username) {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	tender, res := getAndValidateTenderByID(w, tenderID)
	if res {
		return
	}
	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithError(w, http.StatusForbidden, "пользователь не имеет ответственности к организации стенда")
	}

	// Формируем ответ со статусом тендера
	response := map[string]string{
		"status": string(tender.Status),
	}

	// Отправляем успешный ответ с текущим статусом
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// готово
func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из URL (tenderId, status, username)
	tenderID := chi.URLParam(r, "tenderId")
	newStatus := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")
	if validateID(w, tenderID, "ID тендера") || validateStatus(w, models.Status(newStatus)) || validateUsername(w, username) {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	// Проверяем, имеет ли пользователь право изменять статус тендера
	if !database.CheckUserOrganizationResponsibility(user.ID, tenderID) {
		respondWithError(w, http.StatusForbidden, "Пользователь не имеет отвественности за организацию текущего тендера для изменения статуса")
		return
	}
	tender, res := getAndValidateTenderByID(w, tenderID)
	if res {
		return
	}
	// Обновляем статус тендера в базе данных
	tender.Status = models.Status(newStatus)
	if err := database.UpdateTender(tender); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса тендера")
		return
	}
	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}

// исправить готово
func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	if validateID(w, tenderID, "ID тендера") || validateUsername(w, username) {
		return
	}
	tender, res := getAndValidateTenderByID(w, tenderID)
	if res {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}
	if !database.CheckUserOrganizationResponsibility(user.ID, tender.OrganizationID) {
		respondWithError(w, http.StatusForbidden, "Недостаточно прав для редактирования тендера")
		return
	}

	var tenderEditRequest models.TenderEditRequest
	err := json.NewDecoder(r.Body).Decode(&tenderEditRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	if tenderEditRequest.Name == "" && tenderEditRequest.Description == "" && tenderEditRequest.ServiceType == "" {
		respondWithError(w, http.StatusBadRequest, "Отправлен пустой запрос")
		return
	}
	if saveTenderHistory(w, tender) {
		return
	}
	tender.Version++

	if updateTenderFields(w, &tenderEditRequest, tender) {
		return
	}

	if err := database.UpdateTender(tender); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}

// готво
func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	tenderID := chi.URLParam(r, "tenderId")
	versionParam := chi.URLParam(r, "version")
	username := r.URL.Query().Get("username")
	if validateID(w, tenderID, "ID тендера") || validateUsername(w, username) {
		return
	}
	tender, res := getAndValidateTenderByID(w, tenderID)
	if res {
		return
	}
	user, res := getAndValidateUserByUsername(w, username)
	if res {
		return
	}

	version, err := strconv.Atoi(versionParam)
	if err != nil || version < 1 || tender.Version <= version {
		respondWithError(w, http.StatusBadRequest, "Некорректная версия")
		return
	}

	if !database.CheckUserOrganizationResponsibility(user.ID, tenderID) {
		respondWithError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
		return
	}

	tenderHistoryVersion, err := database.GetTenderHistoryByVersion(tenderID, version)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Версия тендера не найдено")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении версии тендера")
		return
	}
	if saveTenderHistory(w, tender) {
		return
	}
	tender.Name = tenderHistoryVersion.Name
	tender.Description = tenderHistoryVersion.Description
	tender.ServiceType = tenderHistoryVersion.ServiceType
	tender.Version++
	if err := database.UpdateTender(tender); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createTenderResponse(tender))
}
