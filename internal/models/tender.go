package models

import "time"

// TenderStatus представляет возможные статусы тендера
type TenderStatus string

const (
	StatusCreated   TenderStatus = "CREATED"   // Тендер создан, но еще не опубликован
	StatusPublished TenderStatus = "PUBLISHED" // Тендер опубликован и доступен для предложений
	StatusClosed    TenderStatus = "CLOSED"    // Тендер закрыт и больше не доступен для предложений
)

// Tender представляет структуру данных для тендера в системе
type Tender struct {
	ID             int          `json:"id"`              // Уникальный идентификатор тендера
	Name           string       `json:"name"`            // Название тендера
	Description    string       `json:"description"`     // Описание тендера
	Status         TenderStatus `json:"status"`          // Статус тендера (например, CREATED, PUBLISHED, CLOSED)
	OrganizationID int          `json:"organization_id"` // ID организации, создавшей тендер
	CreatorID      int          `json:"creator_id"`      // ID пользователя, создавшего тендер
	Version        int          `json:"version"`         // Версия тендера для откатов и обновлений
	CreatedAt      time.Time    `json:"created_at"`      // Дата создания тендера
	UpdatedAt      time.Time    `json:"updated_at"`      // Дата последнего обновления тендера
}
	