package models

import "time"

// OrganizationType represents the type of the organization (IE, LLC, JSC)
type OrganizationType string

const (
	IE  OrganizationType = "IE"  // Индивидуальный предприниматель
	LLC OrganizationType = "LLC" // Общество с ограниченной ответственностью
	JSC OrganizationType = "JSC" // Акционерное общество
)

// Organization represents an organization entity in the system
type Organization struct {
	ID          string           `json:"id"`          // Уникальный идентификатор организации
	Name        string           `json:"name"`        // Название организации
	Description string           `json:"description"` // Описание организации
	Type        OrganizationType `json:"type"`        // Тип организации (IE, LLC, JSC)
	CreatedAt   time.Time        `json:"created_at"`  // Дата создания организации
	UpdatedAt   time.Time        `json:"updated_at"`  // Дата последнего обновления
}
