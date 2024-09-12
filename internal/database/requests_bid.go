package database

import (
	"context"
	"fmt"
	"tender-service/internal/models"
	"time"

	"github.com/jackc/pgx/v4"
)

func EmployeeExists(employeeID string) (bool, error) {
	// Устанавливаем тайм-аут на 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Отмена контекста после завершения функции

	query := `SELECT EXISTS(SELECT 1 FROM employee WHERE id = $1)`

	var exists bool
	err := dbConn.QueryRow(ctx, query, employeeID).Scan(&exists)

	return exists, err
}

func OrganizationExists(organizationID string) (bool, error) {
	// Устанавливаем тайм-аут на 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Отмена контекста после завершения функции

	query := `SELECT EXISTS(SELECT 1 FROM organization WHERE id = $1)`

	var exists bool
	// Выполняем запрос с тайм-аутом
	err := dbConn.QueryRow(ctx, query, organizationID).Scan(&exists)

	return exists, err
}

// починить..
func SaveBid(bid *models.Bid) error {
	// Создаем контекст с тайм-аутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Отмена контекста после завершения функции

	query := `
        INSERT INTO bids (id, name, description, status, tender_id, author_type, author_id, organization_id, version, created_at)
        VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
        RETURNING id, created_at
    `
	// Выполняем запрос и захватываем автоматически сгенерированные поля id и created_at
	err := dbConn.QueryRow(ctx, query, bid.Name, bid.Description, bid.Status, bid.TenderID, bid.AuthorType, bid.AuthorID, bid.OrganizationID, bid.Version).
		Scan(&bid.ID, &bid.CreatedAt)

	return err
}

// версия 1
func GetBidByID(bidID string) (*models.Bid, error) {
	var bid models.Bid

	// Создаем контекст с тайм-аутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Запрос на получение основных данных предложения
	query := `
		SELECT id, name, description, status, tender_id, author_type, author_id, organization_id, version, created_at
		FROM bids
		WHERE id = $1
	`

	// Выполняем запрос и заполняем данные основной структуры Bid
	err := dbConn.QueryRow(ctx, query, bidID).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderID,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.OrganizationID,
		&bid.Version,
		&bid.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Получение всех отзывов (Feedback) для предложения
	feedbackQuery := `
		SELECT id, user_id, bid_id, feedback, created_at
		FROM feedback
		WHERE bid_id = $1
	`
	rows, err := dbConn.Query(ctx, feedbackQuery, bidID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении отзывов: %v", err)
	}
	defer rows.Close()

	var feedbacks []models.Feedback
	for rows.Next() {
		var feedback models.Feedback
		err := rows.Scan(&feedback.ID, &feedback.UserID, &feedback.BidID, &feedback.BidFeedback, &feedback.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка при обработке отзыва: %v", err)
		}
		feedbacks = append(feedbacks, feedback)
	}
	bid.Feedback = feedbacks

	// Получение решений пользователей (UserDecision) по предложению
	decisionQuery := `
		SELECT id, user_id, bid_id, decision
		FROM user_decisions
		WHERE bid_id = $1
	`
	decisionRows, err := dbConn.Query(ctx, decisionQuery, bidID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении решений пользователей: %v", err)
	}
	defer decisionRows.Close()

	var userDecisions []models.UserDecision
	for decisionRows.Next() {
		var userDecision models.UserDecision
		err := decisionRows.Scan(&userDecision.ID, &userDecision.UserID, &userDecision.BidID, &userDecision.Decision)
		if err != nil {
			return nil, fmt.Errorf("ошибка при обработке решения пользователя: %v", err)
		}
		userDecisions = append(userDecisions, userDecision)
	}
	bid.UserDecision = userDecisions

	return &bid, nil
}

// GetBidsByUsername возвращает список предложений пользователя с поддержкой пагинации
func GetBidsByUsername(username string, limit, offset int) ([]models.BidResponse, error) {
	var bids []models.BidResponse

	// Создаем контекст с тайм-аутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE author_id = (SELECT id FROM employee WHERE username = $1)
        ORDER BY name
        LIMIT $2 OFFSET $3
    `

	// Выполняем запрос с использованием контекста
	rows, err := dbConn.Query(ctx, query, username, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid models.BidResponse
		if err := rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &bid.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	// Проверяем ошибки после завершения итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bids, nil
}

func GetBidsByTenderID(tenderID string, limit, offset int) ([]models.BidResponse, error) {
	var bids []models.BidResponse

	// Создаем контекст с тайм-аутом 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE tender_id = $1
        ORDER BY name
        LIMIT $2 OFFSET $3
    `

	rows, err := dbConn.Query(ctx, query, tenderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid models.BidResponse
		if err := rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &bid.CreatedAt); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}

	// Проверяем ошибки после завершения итерации
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bids, nil
}

func GetUserOrganization(userID string) (string, bool, error) {
	var organizationID string

	query := `
        SELECT organization_id 
        FROM organization_responsible 
        WHERE user_id = $1
        LIMIT 1
    `

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, userID).Scan(&organizationID)

	if err != nil {
		if err == pgx.ErrNoRows {
			// Если записи не найдено, возвращаем false
			return "", false, nil
		}
		// Возвращаем ошибку, если что-то пошло не так
		return "", false, err
	}

	// Если организация найдена, возвращаем её ID и true
	return organizationID, true, nil
}

// версия 1
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee
		WHERE username = $1
	`

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, username).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func CheckUserDecisionExists(userDecision *models.UserDecision) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM bid_decisions WHERE bid_id = $1 AND user_id = $2
		)
	`
	dbConn.QueryRow(context.Background(), query, userDecision.BidID, userDecision.UserID).Scan(&exists)
	return exists
}

func SaveUserDecision(userDecision *models.UserDecision) error {
	query := `
		INSERT INTO bid_decisions (id, bid_id, user_id, decision, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, created_at
	`
	err := dbConn.QueryRow(context.Background(), query, userDecision.BidID, userDecision.UserID, userDecision.Decision).
		Scan(&userDecision.ID, &userDecision.Created_at)
	return err
}
// func SaveBidDecision(bidID, userID string, decision models.Сoordination) (*models.UserDecision, error) {
// 	// Начинаем транзакцию для сохранения решения
// 	tx, err := dbConn.Begin(context.Background())
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка начала транзакции: %v", err)
// 	}
// 	defer tx.Rollback(context.Background()) // Откат транзакции в случае ошибки

// 	// Проверяем, существует ли запись о решении пользователя по данному предложению
// 	queryCheck := `
// 		SELECT EXISTS (
// 			SELECT 1 FROM bid_decisions WHERE bid_id = $1 AND user_id = $2
// 		)
// 	`
// 	var exists bool
// 	err = tx.QueryRow(context.Background(), queryCheck, bidID, userID).Scan(&exists)
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка при проверке существования решения: %v", err)
// 	}

// 	if exists {
// 		// Если решение уже существует, обновляем его
// 		queryUpdate := `
// 			UPDATE bid_decisions
// 			SET decision = $1, created_at = CURRENT_TIMESTAMP
// 			WHERE bid_id = $2 AND user_id = $3
// 		`
// 		_, err = tx.Exec(context.Background(), queryUpdate, decision, bidID, userID)
// 		if err != nil {
// 			return nil, fmt.Errorf("ошибка при обновлении решения: %v", err)
// 		}
// 	} else {
// 		// Если решения нет, добавляем новую запись
// 		queryInsert := `
// 			INSERT INTO bid_decisions (id, bid_id, user_id, decision, created_at)
// 			VALUES (uuid_generate_v4(), $1, $2, $3, CURRENT_TIMESTAMP)
// 		`
// 		_, err = tx.Exec(context.Background(), queryInsert, bidID, userID, decision)
// 		if err != nil {
// 			return nil, fmt.Errorf("ошибка при сохранении решения: %v", err)
// 		}
// 	}

// 	// Извлекаем запись о решении для возврата
// 	querySelect := `
// 		SELECT id, bid_id, user_id, decision, created_at
// 		FROM bid_decisions
// 		WHERE bid_id = $1 AND user_id = $2
// 	`
// 	var userDecision models.UserDecision
// 	err = tx.QueryRow(context.Background(), querySelect, bidID, userID).Scan(
// 		&userDecision.ID,
// 		&userDecision.BidID,
// 		&userDecision.UserID,
// 		&userDecision.Decision,
// 		&userDecision.Created_at,
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка при извлечении данных решения: %v", err)
// 	}

// 	// Завершаем транзакцию
// 	err = tx.Commit(context.Background())
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка завершения транзакции: %v", err)
// 	}

// 	return &userDecision, nil
// }

func GetBidsByTenderIDWithExpectation(tenderID string) ([]models.Bid, error) {
	// Создаем список для хранения найденных предложений
	var bids []models.Bid

	// SQL-запрос для поиска предложений по tender_id и состоянию Expectation
	query := `
		SELECT id, name, description, status, tender_id, author_type, author_id, organization_id, version, coordination, created_at
		FROM bids
		WHERE tender_id = $1 AND coordination = $2
	`

	// Выполняем запрос
	rows, err := dbConn.Query(context.Background(), query, tenderID, models.Expectation)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении предложений: %v", err)
	}
	defer rows.Close()

	// Обрабатываем результаты запроса
	for rows.Next() {
		var bid models.Bid
		err := rows.Scan(
			&bid.ID,
			&bid.Name,
			&bid.Description,
			&bid.Status,
			&bid.TenderID,
			&bid.AuthorType,
			&bid.AuthorID,
			&bid.OrganizationID,
			&bid.Version,
			&bid.Сoordination,
			&bid.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании предложения: %v", err)
		}
		bids = append(bids, bid)
	}

	// Проверяем на наличие ошибок после завершения итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке предложений: %v", err)
	}

	return bids, nil
}

// UpdateBid обновляет все данные предложения по его ID
func UpdateBid(bid *models.Bid) error {
	query := `
		UPDATE bids 
		SET name = $1, description = $2, status = $3, tender_id = $4, 
		    author_type = $5, author_id = $6, organization_id = $7, 
		    version = $8, coordination = $9
		WHERE id = $10
	`
	_, err := dbConn.Exec(context.Background(), query,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderID,
		bid.AuthorType,
		bid.AuthorID,
		bid.OrganizationID,
		bid.Version,
		bid.Сoordination,
		bid.ID,
	)
	return err
}
