package user

// Input structures for user operations

// RegisterInput represents input for user registration
type RegisterInput struct {
	Email            string `json:"email"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	MagicLinkEnabled bool   `json:"magic_link_enabled"`
}

// LoginInput represents input for user login
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUserInput represents input for creating a user (admin operation)
type CreateUserInput struct {
	Email            string `json:"email"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Role             string `json:"role"`
	MagicLinkEnabled bool   `json:"magic_link_enabled"`
}

// UpdateProfileInput represents input for updating user profile
type UpdateProfileInput struct {
	Username         *string `json:"username,omitempty"`
	FirstName        *string `json:"first_name,omitempty"`
	LastName         *string `json:"last_name,omitempty"`
	MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}

// AdminUpdateUserInput represents input for admin user updates
type AdminUpdateUserInput struct {
	Username         *string `json:"username,omitempty"`
	FirstName        *string `json:"first_name,omitempty"`
	LastName         *string `json:"last_name,omitempty"`
	Role             *string `json:"role,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
	EmailVerified    *bool   `json:"email_verified,omitempty"`
	MagicLinkEnabled *bool   `json:"magic_link_enabled,omitempty"`
}

// ChangePasswordInput represents input for changing password
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
