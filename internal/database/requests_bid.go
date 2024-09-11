package database

import (
	"context"
	"tender-service/internal/models"
	"time"
	// "fmt"
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
        if err := rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID,  &bid.Version, &bid.CreatedAt); err != nil {
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