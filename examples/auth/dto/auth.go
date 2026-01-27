package dto

// LoginRequest is the request body for login endpoint.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the response body for login endpoint.
type LoginResponse struct {
	Success  bool   `json:"success"`
	Redirect string `json:"redirect,omitempty"`
	Error    string `json:"error,omitempty"`
}

// LogoutResponse is the response body for logout endpoint.
type LogoutResponse struct {
	Success  bool   `json:"success"`
	Redirect string `json:"redirect,omitempty"`
}
