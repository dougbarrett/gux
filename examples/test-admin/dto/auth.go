package dto

// LoginRequest is the request body for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response for login.
type LoginResponse struct {
	Success  bool   `json:"success"`
	Redirect string `json:"redirect,omitempty"`
	Error    string `json:"error,omitempty"`
}

// LogoutResponse is the response for logout.
type LogoutResponse struct {
	Success  bool   `json:"success"`
	Redirect string `json:"redirect,omitempty"`
}
