package models

import "time"

type AuthorType string
type Status string

const (
	AuthorTypeOrganization AuthorType = "Organization"
	AuthorTypeUser         AuthorType = "User"
)

const (
	Created   Status = "Created"
	Published Status = "Published"
	Canceled  Status = "Canceled"
)


type Bid struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	TenderID    string     `json:"tenderId"`
	AuthorType  AuthorType `json:"authorType"`
	AuthorID    string     `json:"authorId"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type BidResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Status      Status     `json:"status"`
	AuthorType  AuthorType `json:"authorType"`
	AuthorID    string     `json:"authorId"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type BidRequest struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	TenderID    string     `json:"tenderId"`
	AuthorType  AuthorType `json:"authorType"`
	AuthorID    string     `json:"authorId"`
}
