package models

// ZerodhaLoginRequest represents the request for Zerodha login
type ZerodhaLoginRequest struct {
	APIKey       string `json:"apiKey" binding:"required"`
	APISecret    string `json:"apiSecret" binding:"required"`
	RequestToken string `json:"requestToken" binding:"required"`
	RedirectURI  string `json:"redirectUri,omitempty"`
}

// ICICILoginRequest represents the request for ICICI Direct login
type ICICILoginRequest struct {
	APIKey       string `json:"apiKey" binding:"required"`
	APISecret    string `json:"apiSecret" binding:"required"`
	RequestToken string `json:"requestToken" binding:"required"`
	RedirectURI  string `json:"redirectUri,omitempty"`
}

// LoginResponse represents the response for login endpoints
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// UserStatus represents the login status for each broker
type UserStatus struct {
	ZerodhaLoggedIn bool `json:"zerodhaLoggedIn"`
	ICICILoggedIn   bool `json:"iciciLoggedIn"`
}

// UserStatusResponse represents the response for user status endpoint
type UserStatusResponse struct {
	Success bool       `json:"success"`
	Data    UserStatus `json:"data"`
	Error   string     `json:"error,omitempty"`
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
