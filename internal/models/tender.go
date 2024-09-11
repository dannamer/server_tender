package models

import "time"

// Tender представляет структуру данных для тендера в системе
type Tender struct {
	ID              string    `json:"id"`               // Уникальный идентификатор тендера (UUID)
	Name            string    `json:"name"`             // Название тендера
	Description     string    `json:"description"`      // Описание тендера
	ServiceType     string    `json:"service_type"`     // Тип услуги, к которой относится тендер
	Status          string    `json:"status"`           // Статус тендера (CREATED, PUBLISHED, CLOSED)
	OrganizationID  string    `json:"organization_id"`  // UUID организации, создавшей тендер
	CreatorUsername string    `json:"creator_username"` // Имя пользователя, создавшего тендер (username)
	Version         int       `json:"version"`          // Версия тендера для откатов и обновлений
	CreatedAt       time.Time `json:"created_at"`       // Дата создания тендера
	UpdatedAt       time.Time `json:"updated_at"`       // Дата последнего обновления тендера
}

type TenderRequest struct {
	Name            string `json:"name"`             // Название тендера
	Description     string `json:"description"`      // Описание тендера
	ServiceType     string `json:"serviceType"`      // Тип услуги, к которой относится тендер
	OrganizationID  string `json:"organizationId"`   // UUID организации, создавшей тендер
	CreatorUsername string `json:"creatorUsername"`  // Имя пользователя, создавшего тендер (username)
}

type TenderResponse struct {
	ID          string    `json:"id"`          // Уникальный идентификатор тендера (UUID)
	Name        string    `json:"name"`        // Название тендера
	Description string    `json:"description"` // Описание тендера
	Status      string    `json:"status"`      // Статус тендера (CREATED, PUBLISHED, CLOSED)
	ServiceType string    `json:"serviceType"` // Тип услуги, к которой относится тендер
	Version     int       `json:"version"`     // Версия тендера для откатов и обновлений
	CreatedAt   time.Time `json:"createdAt"`   // Дата создания тендера
}
