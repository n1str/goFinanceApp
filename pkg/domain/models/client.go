package models

import (
	"gorm.io/gorm"
)

// Клиент системы (бывшая модель User)
type Client struct {
	gorm.Model
	FullName  string   `json:"fullName" gorm:"not null"`
	LoginName string   `json:"loginName" gorm:"unique;not null"`
	Contact   string   `json:"contact" gorm:"unique;not null"`
	PassHash  string   `json:"-" gorm:"not null"`
	Grants    []Access `json:"grants" gorm:"many2many:client_access;"`
}

// ИмеетРазрешение проверяет, имеет ли клиент указанное разрешение
func (c *Client) HasPermission(permissionName string) bool {
	for _, access := range c.Grants {
		if access.Name == permissionName {
			return true
		}
	}
	return false
}

// ЯвляетсяАдминистратором проверяет, имеет ли клиент права администратора
func (c *Client) IsAdministrator() bool {
	return c.HasPermission(AccessAdmin)
}
