package models

import (
	"time"
)

type ServiceType string

const (
	Construction ServiceType = "Construction"
	Delivery     ServiceType = "Delivery"
	Manufacture  ServiceType = "Manufacture"
)

// версия 1
type Tender struct {
	ID                string
	Name              string
	Description       string
	ServiceType       ServiceType
	Status            Status
	OrganizationID    string
	CreatorUsernameID string
	Version           int
	CreatedAt         time.Time
}

type TenderHistory struct {
	ID          string
	TenderID    string
	Name        string
	Description string
	ServiceType ServiceType
	Version     int
}

type TenderRequest struct {
	Name            string      `json:"name"`            // Название тендера
	Description     string      `json:"description"`     // Описание тендера
	ServiceType     ServiceType `json:"serviceType"`     // Тип услуги, к которой относится тендер
	OrganizationID  string      `json:"organizationId"`  // UUID организации, создавшей тендер
	CreatorUsername string      `json:"creatorUsername"` // Имя пользователя, создавшего тендер (username)
}

type TenderResponse struct {
	ID             string      `json:"id"`          // Уникальный идентификатор тендера (UUID)
	Name           string      `json:"name"`        // Название тендера
	Description    string      `json:"description"` // Описание тендера
	ServiceType    ServiceType `json:"serviceType"` // Тип услуги, к которой относится тендер
	Status         Status      `json:"status"`      // Статус тендера (CREATED, PUBLISHED, CLOSED)
	OrganizationID string      `json:"organizationId"`
	Version        int         `json:"version"`   // Версия тендера для откатов и обновлений
	CreatedAt      string      `json:"createdAt"` // Дата создания тендера
}

type TenderEditRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	ServiceType ServiceType `json:"serviceType"`
}
