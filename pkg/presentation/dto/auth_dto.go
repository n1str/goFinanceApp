package dto

// RegisterRequest представляет запрос на регистрацию нового клиента
type RegisterRequest struct {
	FullName  string `json:"fullName" binding:"required"`
	LoginName string `json:"loginName" binding:"required"`
	Contact   string `json:"contact" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// LoginRequest представляет запрос на вход в систему
type LoginRequest struct {
	LoginName string `json:"loginName" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// AuthResponse представляет ответ на успешную аутентификацию
type AuthResponse struct {
	Token     string      `json:"token"`
	Message   string      `json:"message"`
	ClientInfo ClientInfo `json:"client"`
}

// ClientInfo представляет базовую информацию о клиенте
type ClientInfo struct {
	ID        uint   `json:"id"`
	FullName  string `json:"fullName"`
	LoginName string `json:"loginName"`
	Contact   string `json:"contact"`
}

// AuthStatusResponse представляет ответ о статусе аутентификации
type AuthStatusResponse struct {
	Authenticated bool       `json:"authenticated"`
	ClientInfo    ClientInfo `json:"client,omitempty"`
}
