package models

type User struct {
	ID           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Role         Role   `json:"role"`
	TokenVersion int    `json:"-"`
}

func NewUser(id string, firstName string, lastName string, email string, role Role, tokenVersion int) User {
	return User{
		ID:           id,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Role:         role,
		TokenVersion: tokenVersion,
	}
}

func (u User) Equal(other User) bool {
	if u.ID != other.ID {
		return false
	} else if u.FirstName != other.FirstName {
		return false
	} else if u.LastName != other.LastName {
		return false
	} else if u.Email != other.Email {
		return false
	} else if u.Role != other.Role {
		return false
	} else if u.TokenVersion != other.TokenVersion {
		return false
	}
	return true
}
