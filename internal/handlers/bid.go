package handlers

import (
	"encoding/json"
	"github.com/jackc/pgx/v4"
	"net/http"
	"strconv"
	"tender-service/internal/database"
	"tender-service/internal/models"
	"github.com/go-chi/chi/v5"
)

func CreateBidHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что запрос - POST
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
		return
	}

	// Декодируем тело запроса
	var bidRequest models.BidRequest
	err := json.NewDecoder(r.Body).Decode(&bidRequest)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	if validateBidRequest(w, &bidRequest) {
		return
	}

	// Проверяем, существует ли тендер с данным ID
	tender, err := database.GetTenderByID(bidRequest.TenderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Тендер не найден")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
		return
	}

	if tender.Status != string(models.Published) {
		respondWithError(w, http.StatusForbidden, "Тендер ещё не опубликован или закрыт")
		return
	}

	// Создаем новое предложение (bid)
	bid := models.Bid{
		Name:        bidRequest.Name,
		Description: bidRequest.Description,
		TenderID:    bidRequest.TenderID,
		AuthorType:  bidRequest.AuthorType,
		AuthorID:    bidRequest.AuthorID,
		Status:      models.Created,
		Version:     1,
	}

	// Сохраняем предложение в базу данных
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
		CreatedAt:  bid.CreatedAt,
	}
	// Возвращаем успешный ответ с информацией о созданном предложении
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bidResponse)
}

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


func GetBidsForTenderHandler(w http.ResponseWriter, r *http.Request) {
    // Проверка метода запроса
    if r.Method != http.MethodGet {
        respondWithError(w, http.StatusMethodNotAllowed, "Метод не поддерживается")
        return
    }

    // Получаем параметры из URL
    tenderID := chi.URLParam(r, "tenderId")
    username := r.URL.Query().Get("username")

    // Проверяем, заданы ли обязательные параметры
    if tenderID == "" || username == "" {
        respondWithError(w, http.StatusBadRequest, "Отсутствуют обязательные параметры: tenderId или username")
        return
    }

    // Получаем параметры пагинации
    limitParam := r.URL.Query().Get("limit")
    offsetParam := r.URL.Query().Get("offset")
    limit := 5  // Значение по умолчанию
    offset := 0 // Значение по умолчанию

    // Парсим limit и offset
    if limitParam != "" {
        if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        } else {
            respondWithError(w, http.StatusBadRequest, "Некорректное значение limit")
            return
        }
    }
    if offsetParam != "" {
        if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        } else {
            respondWithError(w, http.StatusBadRequest, "Некорректное значение offset")
            return
        }
    }

	userExists, err := database.CheckUserExists(username)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Ошибка проверки пользователя")
        return
    }

    if !userExists {
        respondWithError(w, http.StatusUnauthorized, "Пользователь не существует")
        return
    }
	
    // Проверяем, существует ли тендер
    tender, err := database.GetTenderByID(tenderID)
    if err != nil {
        if err == pgx.ErrNoRows {
            respondWithError(w, http.StatusNotFound, "Тендер не найден")
            return
        }
        respondWithError(w, http.StatusInternalServerError, "Ошибка при получении тендера")
        return
    }

    // Проверяем права пользователя
    if !database.CheckUserOrganizationResponsibility(username, tender.OrganizationID) {
        respondWithError(w, http.StatusForbidden, "Недостаточно прав для выполнения действия")
        return
    }


    // Получаем список предложений для указанного тендера
    bids, err := database.GetBidsByTenderID(tenderID, limit, offset)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Ошибка при получении предложений")
        return
    }

    // Если предложений нет, возвращаем 404
    if len(bids) == 0 {
        respondWithError(w, http.StatusNotFound, "Предложения не найдены")
        return
    }

    // Возвращаем список предложений
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(bids)
}

