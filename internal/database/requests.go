package database

import (
	"context"
	"log"
	"tender-service/internal/models"
	"time"

	"github.com/lib/pq"
)

func CheckUserOrganizationResponsibility(userID string, organizationID string) bool {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible AS org
			WHERE org.user_id = $1 AND org.organization_id = $2
		)
	`

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, userID, organizationID).Scan(&exists)
	if err != nil {
		return false
	}
	// Возвращаем результат проверки
	return exists
}

func SaveTender(tender *models.Tender) error {
	query := `
		INSERT INTO tenders (id, name, description, service_type, status, organization_id, creator_username_id, version, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		RETURNING id, created_at
	`

	err := dbConn.QueryRow(context.Background(), query,
		tender.Name,
		tender.Description,
		tender.ServiceType,
		tender.Status,
		tender.OrganizationID,
		tender.CreatorUsernameID,
		tender.Version,
	).Scan(&tender.ID, &tender.CreatedAt)

	return err
}

// GetTenders возвращает список тендеров с учетом фильтров по типам услуг, лимита и смещения
func GetTendersResponse(serviceTypes []string, limit, offset int) ([]models.TenderResponse, error) {
	// Формируем базовый SQL-запрос
	query := `
		SELECT id, name, description, service_type, status, version, created_at 
		FROM tenders
	`

	// Добавляем условие для фильтрации по типам услуг, если указано
	var args []interface{}
	if len(serviceTypes) > 0 {
		query += " WHERE service_type = ANY($1)"
		args = append(args, pq.Array(serviceTypes))
	}

	// Добавляем сортировку и параметры пагинации
	query += " ORDER BY name LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	// Выполняем запрос к базе данных
	rows, err := dbConn.Query(context.Background(), query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()

	// Обрабатываем результат и заполняем список тендеров
	var tenders []models.TenderResponse
	for rows.Next() {
		var tender models.TenderResponse
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Printf("Ошибка при обработке результата запроса: %v", err)
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	// Проверяем наличие ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при чтении строк: %v", err)
		return nil, err
	}

	return tenders, nil
}

func GetTendersByUsername(username string, limit, offset int) ([]models.TenderResponse, error) {
	query := `
		SELECT id, name, description, service_type, status, version, created_at
		FROM tenders
		WHERE creator_username = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := dbConn.Query(context.Background(), query, username, limit, offset)
	if err != nil {
		log.Printf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tenders []models.TenderResponse
	for rows.Next() {
		var tender models.TenderResponse
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &tender.CreatedAt)
		if err != nil {
			log.Printf("Ошибка при обработке результата запроса: %v", err)
			return nil, err
		}
		tenders = append(tenders, tender)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Ошибка при чтении строк: %v", err)
		return nil, err
	}

	return tenders, nil
}

// версия 1
func GetTenderByID(tenderID string) (*models.Tender, error) {
	var tender models.Tender

	// Устанавливаем контекст с тайм-аутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// SQL-запрос для получения тендера по ID
	query := `
		SELECT id, name, description, service_type, status, organization_id, creator_username_id, version, created_at
		FROM tenders
		WHERE id = $1
	`

	// Выполняем запрос и сканируем результат в структуру Tender
	err := dbConn.QueryRow(ctx, query, tenderID).Scan(
		&tender.ID,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.OrganizationID,
		&tender.CreatorUsernameID,
		&tender.Version,
		&tender.CreatedAt,
	)

	// Если произошла ошибка, возвращаем ее
	if err != nil {
		return nil, err
	}

	// Возвращаем структуру Tender и nil как ошибку
	return &tender, nil
}

func UpdateTender(tender *models.Tender) error {
	query := `
		UPDATE tenders 
		SET name = $1, description = $2, service_type = $3, status = $4, 
		    organization_id = $5, creator_username_id = $6, version = $7
		WHERE id = $8
	`
	_, err := dbConn.Exec(context.Background(), query,
		tender.Name,
		tender.Description,
		tender.ServiceType,
		tender.Status,
		tender.OrganizationID,
		tender.CreatorUsernameID,
		tender.Version,
		tender.ID,
	)
	return err
}

func SaveTenderHistory(tenderHistory *models.TenderHistory) error {
	// Запрос на вставку данных в таблицу tender_history
	query := `
		INSERT INTO tender_history (id, tender_id, name, description, service_type, version)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
	`

	// Выполняем запрос с параметрами из структуры TenderHistory
	_, err := dbConn.Exec(context.Background(), query,
		tenderHistory.TenderID,
		tenderHistory.Name,
		tenderHistory.Description,
		tenderHistory.ServiceType,
		tenderHistory.Version,
	)
	return err
}

func GetTenderHistoryByVersion(tenderID string, version int) (*models.TenderHistory, error) {
	var tenderHistory models.TenderHistory
	query := `
		SELECT id, tender_id, name, description, service_type, version
		FROM tender_history
		WHERE tender_id = $1 AND version = $2
	`
	err := dbConn.QueryRow(context.Background(), query, tenderID, version).Scan(
		&tenderHistory.ID,
		&tenderHistory.TenderID,
		&tenderHistory.Name,
		&tenderHistory.Description,
		&tenderHistory.ServiceType,
		&tenderHistory.Version,
	)
	return &tenderHistory, err
}
