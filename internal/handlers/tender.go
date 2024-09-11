package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"net/http"
	"strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
	"time"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		http.Error(w, "Unable to write response", http.StatusInternalServerError)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.ErrorResponse{Reason: message})
}

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

	// Проверяем обязательные поля и ограничения по длине
	if len(tenderRequest.Name) > 100 || len(tenderRequest.Description) > 500 {
		respondWithError(w, http.StatusBadRequest, "Превышены ограничения по длине полей")
		return
	}

	// Проверяем формат UUID для organizationId
	if _, err := uuid.Parse(tenderRequest.OrganizationID); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный UUID для organizationId")
		return
	}

	// Проверяем обязательные поля
	if tenderRequest.Name == "" || tenderRequest.Description == "" || tenderRequest.ServiceType == "" ||  tenderRequest.OrganizationID == "" || tenderRequest.CreatorUsername == "" {
		respondWithError(w, http.StatusBadRequest, "Отсутствуют обязательные поля")
		return
	}

	// Проверяем права пользователя
	isResponsible := database.CheckUserOrganizationResponsibility(tenderRequest.CreatorUsername, tenderRequest.OrganizationID)
	if !isResponsible {
		respondWithError(w, http.StatusForbidden, "Пользователь не имеет прав для создания тендеров от имени этой организации")
		return
	}

	// Создаем объект тендера
	tender := models.Tender{
		Name:            tenderRequest.Name,
		Description:     tenderRequest.Description,
		ServiceType:     tenderRequest.ServiceType,
		Status:          "Created",
		OrganizationID:  tenderRequest.OrganizationID,
		CreatorUsername: tenderRequest.CreatorUsername,
		Version:         1,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Сохраняем тендер в базу данных
	err = database.SaveTender(&tender)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при сохранении тендера")
		return
	}

	// Формируем ответ с нужными полями
	response := models.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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
	tenders, err := database.GetTenders(serviceTypes, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендеров")
		return
	}

	// Возвращаем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tenders)
}

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

func GetTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из URL (tenderId и username)
	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	if tenderID == "" {
		respondWithError(w, http.StatusBadRequest, "tenderId обязателен")
		return
	}

	if username == "" {
		respondWithError(w, http.StatusUnauthorized, "username обязателен")
		return
	}

	// Проверяем, существует ли тендер с данным ID
	tender, err := database.GetTenderByID(tenderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Тендер не найден")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
		return
	}

	// Формируем ответ со статусом тендера
	response := map[string]string{
		"status": tender.Status,
	}

	// Отправляем успешный ответ с текущим статусом
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UpdateTenderStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из URL (tenderId, status, username)
	tenderID := chi.URLParam(r, "tenderId")
	newStatus := r.URL.Query().Get("status")
	username := r.URL.Query().Get("username")

	if tenderID == "" || newStatus == "" || username == "" {
		respondWithError(w, http.StatusBadRequest, "Все параметры (tenderId, status, username) обязательны")
		return
	}

	// Проверка допустимых статусов
	validStatuses := map[string]bool{"Created": true, "Published": true, "Closed": true}
	if !validStatuses[newStatus] {
		respondWithError(w, http.StatusBadRequest, "Некорректный статус")
		return
	}

	// Проверяем, имеет ли пользователь право изменять статус тендера
	isResponsible := database.CheckUserOrganizationResponsibility(username, tenderID)
	if !isResponsible {
		respondWithError(w, http.StatusForbidden, "Пользователь не имеет прав для изменения статуса тендера")
		return
	}

	// Обновляем статус тендера в базе данных
	err := database.UpdateTenderStatus(tenderID, newStatus)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении статуса тендера")
		return
	}
	tender, _ := database.GetTenderByID(tenderID)
	response := models.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}
	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func EditTenderHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPatch {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из URL (tenderId и username)
	tenderID := chi.URLParam(r, "tenderId")
	username := r.URL.Query().Get("username")

	userExists, err := database.CheckUserExists(username)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при проверке пользователя")
		return
	}
	if !userExists {
		respondWithError(w, http.StatusUnauthorized, "Пользователь не существует или некорректен")
		return
	}

	// Проверяем наличие tenderId и username
	if tenderID == "" || username == "" {
		respondWithError(w, http.StatusBadRequest, "tenderId и username обязательны")
		return
	}

	// Проверяем, существует ли тендер с данным ID
	tender, err := database.GetTenderByID(tenderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Тендер не найден")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
		return
	}

	// Проверяем, имеет ли пользователь права на редактирование тендера
	isResponsible := database.CheckUserOrganizationResponsibility(username, tender.OrganizationID)
	if !isResponsible {
		respondWithError(w, http.StatusForbidden, "Недостаточно прав для редактирования тендера")
		return
	}

	// Декодируем тело запроса
	var updateRequest models.TenderRequest
	err = json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат запроса")
		return
	}

	// Обновляем только те поля, которые были переданы
	if updateRequest.Name != "" {
		tender.Name = updateRequest.Name
	}
	if updateRequest.Description != "" {
		tender.Description = updateRequest.Description
	}
	if updateRequest.ServiceType != "" {
		tender.ServiceType = updateRequest.ServiceType
	}

	// Обновляем время последнего изменения
	tender.UpdatedAt = time.Now()

	// Сохраняем обновленный тендер в базу данных
	err = database.UpdateTender(tender)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
		return
	}

	// Возвращаем обновленный тендер в ответ
	response := models.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		ServiceType: tender.ServiceType,
		Status:      tender.Status,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func RollbackTenderHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodPut {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Получаем параметры из пути
	tenderID := chi.URLParam(r, "tenderId")
	versionParam := chi.URLParam(r, "version")
	username := r.URL.Query().Get("username")

	// Проверяем, переданы ли обязательные параметры
	if tenderID == "" || versionParam == "" || username == "" {
		respondWithError(w, http.StatusBadRequest, "Отсутствуют обязательные параметры")
		return
	}

	// Преобразуем версию в int
	version, err := strconv.Atoi(versionParam)
	if err != nil || version < 1 {
		respondWithError(w, http.StatusBadRequest, "Некорректная версия")
		return
	}

	// Проверка прав пользователя (например, ответственен ли пользователь за тендер)
	if !database.CheckUserOrganizationResponsibility(username, tenderID) {
		respondWithError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
		return
	}

	// Получаем нужную версию тендера из таблицы
	tenderHistory, err := database.GetTenderHistoryByVersion(tenderID, version)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Тендер или версия не найдены")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении версии тендера")
		return
	}

	tenderHistory.Version++

	// Обновляем тендер до новой версии
	err = database.UpdateTenderFully(tenderHistory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка при обновлении тендера")
		return
	}

	// Формируем ответ с обновленными данными тендера
	response := models.TenderResponse{
		ID:          tenderHistory.ID,
		Name:        tenderHistory.Name,
		Description: tenderHistory.Description,
		Status:      tenderHistory.Status,
		ServiceType: tenderHistory.ServiceType,
		Version:     tenderHistory.Version,
		CreatedAt:   tenderHistory.CreatedAt,
	}

	// Возвращаем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
