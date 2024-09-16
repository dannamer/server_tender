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
	err := dbConn.QueryRow(context.Background(), query, userID, organizationID).Scan(&exists)
	if err != nil {
		return false
	}
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

func GetTendersResponse(serviceTypes []string, limit, offset int) ([]models.TenderResponse, error) {
	query := `
		SELECT id, name, description, service_type, status, version, created_at 
		FROM tenders
	`
	var args []interface{}
	if len(serviceTypes) > 0 {
		query += " WHERE service_type = ANY($1)"
		args = append(args, pq.Array(serviceTypes))
		query += " ORDER BY name LIMIT $2 OFFSET $3"
	} else {
		query += " ORDER BY name LIMIT $1 OFFSET $2"
	}
	args = append(args, limit, offset)

	rows, err := dbConn.Query(context.Background(), query, args...)
	if err != nil {
		log.Printf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()
	tenders := []models.TenderResponse{}
	for rows.Next() {
		var tender models.TenderResponse
		var created_at time.Time
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &created_at)
		tender.CreatedAt = created_at.Format(time.RFC3339)
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

func GetTendersByUsername(usernameID string, limit, offset int) ([]models.TenderResponse, error) {
	query := `
		SELECT id, name, description, service_type, status, version, created_at
		FROM tenders
		WHERE creator_username_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := dbConn.Query(context.Background(), query, usernameID, limit, offset)
	if err != nil {
		log.Printf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()

	tenders := []models.TenderResponse{}
	for rows.Next() {
		var tender models.TenderResponse
		var createdAt time.Time
		err := rows.Scan(&tender.ID, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.Version, &createdAt)
		tender.CreatedAt = createdAt.Format(time.RFC3339)
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
	query := `
		SELECT id, name, description, service_type, status, organization_id, creator_username_id, version, created_at
		FROM tenders
		WHERE id = $1
	`
	err := dbConn.QueryRow(context.Background(), query, tenderID).Scan(
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
	if err != nil {
		return nil, err
	}

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
	query := `
		INSERT INTO tender_history (id, tender_id, name, description, service_type, version)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5)
	`
	_, err := dbConn.Exec(context.Background(), query,
		tenderHistory.TenderID,
		tenderHistory.Name,
		tenderHistory.Description,
		tenderHistory.ServiceType,
		tenderHistory.Version,
	)
	return err
}

func GetBidHistoryByVersion(bidID string, version int) (*models.BidHistory, error) {
	var bidHistory models.BidHistory
	query := `
		SELECT id, bid_id, name, description, version
		FROM bid_history
		WHERE bid_id = $1 AND version = $2
	`
	err := dbConn.QueryRow(context.Background(), query, bidID, version).Scan(
		&bidHistory.ID,
		&bidHistory.BidID,
		&bidHistory.Name,
		&bidHistory.Description,
		&bidHistory.Version,
	)
	return &bidHistory, err
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


// func GetUserByID(userID string) (*models.User, error) {
// 	var user models.User

// 	query := `
// 		SELECT id, username, first_name, last_name, created_at, updated_at
// 		FROM employee
// 		WHERE id = $1
// 	`

// 	// Выполняем запрос к базе данных
// 	err := dbConn.QueryRow(context.Background(), query, userID).Scan(
// 		&user.ID,
// 		&user.Username,
// 		&user.FirstName,
// 		&user.LastName,
// 		&user.CreatedAt,
// 		&user.UpdatedAt,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &user, nil
// }