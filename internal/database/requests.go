package database

import (
	"context"
	"log"
	"tender-service/internal/models"
)

// Функция проверяет, является ли пользователь ответственным за организацию
func CheckUserOrganizationResponsibility(userID int, organizationID int) bool {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible 
			WHERE user_id = $1 AND organization_id = $2
		)
	`

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, userID, organizationID).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке прав пользователя: %v\n", err)
		return false
	}

	// Возвращаем результат проверки
	return exists
}

func SaveTender(tender *models.Tender) error {
	query := `
		INSERT INTO tenders (name, description, status, organization_id, creator_id, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	// Выполняем запрос и сохраняем id созданного тендера
	err := dbConn.QueryRow(context.Background(), query,
		tender.Name,
		tender.Description,
		string(tender.Status),
		tender.OrganizationID,
		tender.CreatorID,
		tender.Version,
		tender.CreatedAt,
		tender.UpdatedAt).Scan(&tender.ID)

	if err != nil {
		log.Printf("Ошибка при сохранении тендера: %v\n", err)
		return err
	}

	log.Println("Тендер успешно сохранён, ID:", tender.ID)
	return nil
}
