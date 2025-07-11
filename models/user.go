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

func (self User) Equal(other User) bool {
	if self.ID != other.ID {
		return false
	} else if self.FirstName != other.FirstName {
		return false
	} else if self.LastName != other.LastName {
		return false
	} else if self.Email != other.Email {
		return false
	} else if self.Role != other.Role {
		return false
	} else if self.TokenVersion != other.TokenVersion {
		return false
	}
	return true
}
