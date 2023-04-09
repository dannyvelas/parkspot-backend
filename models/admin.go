package models

type Admin struct {
	ID           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Password     string `json:"-"`
	IsPrivileged bool   `json:"isPrivileged"`
	TokenVersion *int   `json:"-"`
}

func NewAdmin(id string, firstName string, lastName string, email string, password string, isPrivileged bool, tokenVersion int) Admin {
	return Admin{
		ID:           id,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Password:     password,
		IsPrivileged: isPrivileged,
		TokenVersion: &tokenVersion,
	}
}

func (a Admin) GetPassword() string {
	return a.Password
}

func (a Admin) AsUser() User {
	return NewUser(a.ID, a.FirstName, a.LastName, a.Email, AdminRole, *a.TokenVersion)
}
