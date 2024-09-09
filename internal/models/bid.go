package models

import "time"

type Bid struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Status          string    `json:"status"`
	TenderID        int       `json:"tender_id"`
	OrganizationID  int       `json:"organization_id"`
	CreatorUsername string    `json:"creator_username"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
