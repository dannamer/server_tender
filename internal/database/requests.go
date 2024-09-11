package database

import (
	"context"
	"errors"
	"log"
	"tender-service/internal/models"
	"time"

	"github.com/lib/pq"
)

// Функция проверяет, является ли пользователь ответственным за организацию
func CheckUserOrganizationResponsibility(username string, organizationID string) bool {
	var exists bool

	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible AS or
			JOIN employee AS e ON or.user_id = e.id
			WHERE e.username = $1 AND or.organization_id = $2
		)
	`

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, username, organizationID).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка при проверке прав пользователя: %v\n", err)
		return false
	}

	// Возвращаем результат проверки
	return exists
}

func SaveTender(tender *models.Tender) error {
	if dbConn == nil {
		return errors.New("нет подключения к базе данных")
	}

	query := `
		INSERT INTO tenders (id, name, description, service_type, status, organization_id, creator_username, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err := dbConn.QueryRow(context.Background(), query,
		tender.ID,
		tender.Name,
		tender.Description,
		tender.ServiceType,
		tender.Status,
		tender.OrganizationID,
		tender.CreatorUsername,
		tender.Version,
		tender.CreatedAt,
		tender.UpdatedAt,
	).Scan(&tender.ID)

	if err != nil {
		log.Printf("Ошибка при сохранении тендера: %v", err)
		return err
	}

	log.Println("Тендер успешно сохранен с ID:", tender.ID)
	return nil
}

// GetTenders возвращает список тендеров с учетом фильтров по типам услуг, лимита и смещения
func GetTenders(serviceTypes []string, limit, offset int) ([]models.TenderResponse, error) {
	// Устанавливаем тайм-аут в 1 минуту
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel() // Освобождаем ресурсы после завершения функции

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

	// Выполняем запрос к базе данных с контекстом и тайм-аутом
	rows, err := dbConn.Query(ctx, query, args...)
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	query := `
		SELECT id, name, description, service_type, status, version, created_at
		FROM tenders
		WHERE creator_username = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := dbConn.Query(ctx, query, username, limit, offset)
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

func GetTenderByID(tenderID string) (*models.Tender, error) {
	var tender models.Tender

	// Устанавливаем тайм-аут на 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Отмена контекста после завершения функции

	query := `SELECT id, name, description, service_type, status, version, created_at FROM tenders WHERE id = $1`

	// Выполняем запрос с тайм-аутом
	err := dbConn.QueryRow(ctx, query, tenderID).Scan(
		&tender.ID,
		&tender.Name,
		&tender.Description,
		&tender.ServiceType,
		&tender.Status,
		&tender.Version,
		&tender.CreatedAt,
	)

	return &tender, err
}


func UpdateTenderStatus(tenderID, status string) error {
	query := `UPDATE tenders SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := dbConn.Exec(context.Background(), query, status, time.Now(), tenderID)
	return err
}

func UpdateTender(tender *models.Tender) error {
	// Начинаем транзакцию для атомарного сохранения изменений и записи в историю
	tx, err := dbConn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background()) // Откат транзакции в случае ошибки

	// Получаем текущую версию тендера
	tenderLast, err := GetTenderByID(tender.ID)
	if err != nil {
		return err
	}

	// Сохраняем текущую версию тендера в таблицу tender_history
	historyQuery := `
		INSERT INTO tender_history (id, tender_id, name, description, service_type, version, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6)
	`
	_, err = tx.Exec(context.Background(), historyQuery, tenderLast.ID, tenderLast.Name, tenderLast.Description, tenderLast.ServiceType, tenderLast.Version, tenderLast.CreatedAt)
	if err != nil {
		return err
	}

	// Инкрементируем версию тендера
	tender.Version = tenderLast.Version + 1

	// Обновляем тендер
	query := `
		UPDATE tender 
		SET name = $1, description = $2, service_type = $3, updated_at = $4, version = $5
		WHERE id = $6
	`
	_, err = tx.Exec(context.Background(), query, tender.Name, tender.Description, tender.ServiceType, tender.UpdatedAt, tender.Version, tender.ID)
	if err != nil {
		return err
	}

	// Завершаем транзакцию
	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func CheckUserExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM employee WHERE username = $1)`
	var exists bool
	err := dbConn.QueryRow(context.Background(), query, username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetTenderHistoryByVersion(tenderID string, version int) (*models.Tender, error) {
	history, _ := GetTenderByID(tenderID)
	query := `
		SELECT name, description, service_type
		FROM tender_history
		WHERE tender_id = $1 AND version = $2
	`
	history.UpdatedAt = time.Now()
	err := dbConn.QueryRow(context.Background(), query, tenderID, version).Scan(&history.Name, &history.Description, &history.ServiceType)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func UpdateTenderFully(tender *models.Tender) error {
	// Запрос на обновление всех полей тендера по его ID
	query := `
		UPDATE tender 
		SET 
			name = $1, 
			description = $2, 
			service_type = $3, 
			status = $4, 
			organization_id = $5, 
			creator_username = $6, 
			version = $7, 
			updated_at = $8 
		WHERE id = $9
	`

	// Выполнение запроса с передачей всех значений
	_, err := dbConn.Exec(context.Background(), query,
		tender.Name,
		tender.Description,
		tender.ServiceType,
		tender.Status,
		tender.OrganizationID,
		tender.CreatorUsername,
		tender.Version,
		tender.UpdatedAt, // Обновляем время последнего изменения
		tender.ID,
	)

	if err != nil {
		return err
	}

	return nil
}
