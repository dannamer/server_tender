package handlers

import (
	"tender-service/internal/models"
	"net/http"
	"tender-service/internal/database"
)

func validateBidRequest(w http.ResponseWriter, bid *models.BidRequest) bool {
	if len(bid.Name) > 100 {
		respondWithError(w, http.StatusBadRequest, "Название предложения слишком длинное, максимум 100 символов")
		return true
	}
	if len(bid.Description) > 500 {
		respondWithError(w, http.StatusBadRequest, "Описание предложения слишком длинное, максимум 500 символов")
		return true
	}
	if len(bid.TenderID) > 100 {
		respondWithError(w, http.StatusBadRequest, "ID тендера слишком длинное, максимум 100 символов")
		return true
	}
	if len(bid.AuthorID) > 100 {
		respondWithError(w, http.StatusBadRequest, "ID автора слишком длинное, максимум 100 символов")
		return true
	}
	if bid.AuthorType != models.AuthorTypeOrganization && bid.AuthorType != models.AuthorTypeUser {
		respondWithError(w, http.StatusBadRequest, "Некорректное значение authorType")
		return true
	}
	if bid.Name == "" || bid.Description == "" || bid.TenderID == "" || bid.AuthorID == "" {
		respondWithError(w, http.StatusBadRequest, "Отсутствуют обязательные поля")
		return true
	}
	if bid.AuthorType == models.AuthorTypeUser {
		exists, err := database.EmployeeExists(bid.AuthorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка проверки существования пользователя")
			return true
		}
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "Пользователь не найден")
			return true
		}
	} else {
		exists, err := database.OrganizationExists(bid.AuthorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка проверки существования организации")
			return true
		}
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "Организация не найдена")
			return true
		}
	}
	return false
}