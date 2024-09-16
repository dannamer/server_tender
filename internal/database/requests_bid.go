package database

import (
	"context"
	"fmt"
	"tender-service/internal/models"
	"time"
)

func SaveBid(bid *models.Bid) error {
	query := `
        INSERT INTO bids (id, name, description, status, tender_id, author_type, author_id, version, coordination, created_at)
        VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
        RETURNING id, created_at
    `
	err := dbConn.QueryRow(context.Background(), query, bid.Name, bid.Description, bid.Status, bid.TenderID, bid.AuthorType, bid.AuthorID, bid.Version, bid.Сoordination).
		Scan(&bid.ID, &bid.CreatedAt)

	return err
}

// версия 1
func GetBidByID(bidID string) (*models.Bid, error) {
	var bid models.Bid
	query := `
		SELECT id, name, description, status, tender_id, author_type, author_id, version, created_at
		FROM bids
		WHERE id = $1
	`
	err := dbConn.QueryRow(context.Background(), query, bidID).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderID,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	feedbackQuery := `
		SELECT id, user_id, bid_id, bid_feedback, created_at
		FROM feedback
		WHERE bid_id = $1
	`
	rows, err := dbConn.Query(context.Background(), feedbackQuery, bidID)
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
	decisionQuery := `
		SELECT id, user_id, bid_id, decision
		FROM user_decisions
		WHERE bid_id = $1
	`
	decisionRows, err := dbConn.Query(context.Background(), decisionQuery, bidID)
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

func GetBidsByUserID(authorID string, limit, offset int) ([]models.BidResponse, error) {
	bids := []models.BidResponse{}
	query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE author_id = $1
        ORDER BY name
        LIMIT $2 OFFSET $3
    `
	rows, err := dbConn.Query(context.Background(), query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid models.BidResponse
		var createdAt time.Time
		if err := rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &createdAt); err != nil {
			return nil, err
		}
		bid.CreatedAt = createdAt.Format(time.RFC3339)
		bids = append(bids, bid)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bids, nil
}

func GetBidsByTenderID(tenderID string, limit, offset int) ([]models.BidResponse, error) {
	bids := []models.BidResponse{}

	query := `
        SELECT id, name, status, author_type, author_id, version, created_at
        FROM bids
        WHERE tender_id = $1
        ORDER BY name
        LIMIT $2 OFFSET $3
    `

	rows, err := dbConn.Query(context.Background(), query, tenderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid models.BidResponse
		var createdAt time.Time
		if err := rows.Scan(&bid.ID, &bid.Name, &bid.Status, &bid.AuthorType, &bid.AuthorID, &bid.Version, &createdAt); err != nil {
			return nil, err
		}
		bid.CreatedAt = createdAt.Format(time.RFC3339)
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bids, nil
}

// версия 1
func GetUserByUsername(username string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee
		WHERE username = $1
	`
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

func SaveFeedback(feedback *models.Feedback) error {
	query := `
		INSERT INTO feedback (id, user_id, bid_id, bid_feedback, created_at)
		VALUES (uuid_generate_v4(), $1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, created_at
	`
	err := dbConn.QueryRow(context.Background(), query, feedback.UserID, feedback.BidID, feedback.BidFeedback).
		Scan(&feedback.ID, &feedback.CreatedAt)
	return err
}

func GetBidsByTenderIDWithExpectation(tenderID string) ([]models.Bid, error) {
	var bids []models.Bid

	query := `
		SELECT id, name, description, status, tender_id, author_type, author_id, version, coordination, created_at
		FROM bids
		WHERE tender_id = $1 AND coordination = $2
	`

	rows, err := dbConn.Query(context.Background(), query, tenderID, models.Expectation)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении предложений: %v", err)
	}
	defer rows.Close()

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
			&bid.Version,
			&bid.Сoordination,
			&bid.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании предложения: %v", err)
		}
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке предложений: %v", err)
	}

	return bids, nil
}

func UpdateBid(bid *models.Bid) error {
	query := `
		UPDATE bids 
		SET name = $1, description = $2, status = $3, tender_id = $4, 
		    author_type = $5, author_id = $6, 
		    version = $7, coordination = $8
		WHERE id = $9
	`
	_, err := dbConn.Exec(context.Background(), query,
		bid.Name,
		bid.Description,
		bid.Status,
		bid.TenderID,
		bid.AuthorType,
		bid.AuthorID,
		bid.Version,
		bid.Сoordination,
		bid.ID,
	)
	return err
}

func GetUserByID(userID string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, username, first_name, last_name, created_at, updated_at
		FROM employee
		WHERE id = $1
	`

	// Выполняем запрос к базе данных
	err := dbConn.QueryRow(context.Background(), query, userID).Scan(
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

func GetOrganizationByID(organizationID string) (*models.Organization, error) {
	var organization models.Organization

	query := `
		SELECT id, name, type, created_at, updated_at
		FROM organization
		WHERE id = $1
	`
	err := dbConn.QueryRow(context.Background(), query, organizationID).Scan(
		&organization.ID,
		&organization.Name,
		&organization.Type,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &organization, nil
}

func SaveBidHistory(bidHistory *models.BidHistory) error {
	query := `
		INSERT INTO bid_history (id, bid_id, name, description, version)
		VALUES (uuid_generate_v4(), $1, $2, $3, $4)
	`
	_, err := dbConn.Exec(context.Background(), query,
		bidHistory.BidID,
		bidHistory.Name,
		bidHistory.Description,
		bidHistory.Version,
	)
	return err
}

func GetFeedbackByBidID(bidID string) ([]models.Feedback, error) {
	query := `
        SELECT id, user_id, bid_id, bid_feedback, created_at
        FROM feedback
        WHERE bid_id = $1
        ORDER BY created_at ASC
    `

	rows, err := dbConn.Query(context.Background(), query, bidID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feedbacks := []models.Feedback{}

	for rows.Next() {
		var feedback models.Feedback
		if err := rows.Scan(&feedback.ID, &feedback.UserID, &feedback.BidID, &feedback.BidFeedback, &feedback.CreatedAt); err != nil {
			return nil, err
		}
		feedbacks = append(feedbacks, feedback)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feedbacks, nil
}

func CheckBidExists(authorID, tenderID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM bids 
			WHERE author_id = $1 AND tender_id = $2
		)
	`
	var exists bool
	err := dbConn.QueryRow(context.Background(), query, authorID, tenderID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func GetBidByTenderAndAuthorID(tenderID, authorID string) (*models.Bid, error) {
	var bid models.Bid
	query := `
		SELECT id, name, description, status, tender_id, author_type, author_id, version, coordination, created_at
		FROM bids
		WHERE tender_id = $1 AND author_id = $2
	`
	err := dbConn.QueryRow(context.Background(), query, tenderID, authorID).Scan(
		&bid.ID,
		&bid.Name,
		&bid.Description,
		&bid.Status,
		&bid.TenderID,
		&bid.AuthorType,
		&bid.AuthorID,
		&bid.Version,
		&bid.Сoordination,
		&bid.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &bid, nil
}

func GetBidReviews(bidID string, limit, offset int) ([]models.FeedbackResponse, error) {
	query := `
		SELECT id, bid_feedback, created_at
		FROM feedback
		WHERE bid_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := dbConn.Query(context.Background(), query, bidID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении отзывов: %v", err)
	}
	defer rows.Close()

	feedbackResponses := []models.FeedbackResponse{}
	for rows.Next() {
		var feedbackResponse models.FeedbackResponse
		var createdAt time.Time
		if err := rows.Scan(&feedbackResponse.ID, &feedbackResponse.Description, &createdAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании отзыва: %v", err)
		}
		// Преобразуем время в формат RFC3339
		feedbackResponse.CreatedAt = createdAt.Format(time.RFC3339)
		feedbackResponses = append(feedbackResponses, feedbackResponse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при чтении строк отзывов: %v", err)
	}

	return feedbackResponses, nil
}

func CheckUserOrganization(userID string) bool {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM organization_responsible 
			WHERE user_id = $1
		)
	`
	var exists bool
	dbConn.QueryRow(context.Background(), query, userID).Scan(&exists)
	return exists
}
