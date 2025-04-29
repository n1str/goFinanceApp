package models

// Константы для прав доступа
const (
	AccessUser  = "user"  // Обычный пользователь
	AccessAdmin = "admin" // Администратор
)

// Access определяет права доступа клиента в системе (бывшая Role)
type Access struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"unique;not null"`
	Description string `json:"description" gorm:"not null"`
}
