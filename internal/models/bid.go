package models

import "time"

type AuthorType string
type Status string
type Сoordination string

const (
	AuthorTypeOrganization AuthorType = "Organization"
	AuthorTypeUser         AuthorType = "User"
)

const (
	Created   Status = "Created"
	Published Status = "Published"
	Closed  Status = "Closed"
)

const (
	Expectation        Сoordination = "Expectation"        // Ожидание решения
	Approved           Сoordination = "Approved"           // Одобрено
	Rejected           Сoordination = "Rejected"           // Отклонено
	RejectedByConflict Сoordination = "RejectedByConflict" // Отклонено из-за одобрения тендера с другим предложением
)

// версия 1
type Bid struct {
	ID             string
	Name           string
	Description    string
	Status         Status
	TenderID       string
	AuthorType     AuthorType
	AuthorID       string
	OrganizationID string
	Version        int
	Feedback       []Feedback
	Сoordination   Сoordination
	UserDecision   []UserDecision
	CreatedAt      time.Time
}

type UserDecision struct {
	ID         string
	UserID     string
	BidID      string
	Decision   Сoordination
	Created_at time.Time
}
type Feedback struct {
	ID          string
	UserID      string
	BidID       string
	BidFeedback string
	CreatedAt   time.Time
}

type BidResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     Status     `json:"status"`
	AuthorType AuthorType `json:"authorType"`
	AuthorID   string     `json:"authorId"`
	Version    int        `json:"version"`
	CreatedAt  string     `json:"createdAt"`
}

type BidRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	TenderID    string     `json:"tenderId"`
	AuthorType  AuthorType `json:"authorType"`
	AuthorID    string     `json:"authorId"`
}
