package models

type User struct {
	Id           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Role         Role   `json:"role"`
	TokenVersion int    `json:"-"`
}

func newUser(id string, firstName string, lastName string, email string, role Role, tokenVersion int) User {
	return User{
		Id:           id,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Role:         role,
		TokenVersion: tokenVersion,
	}
}

func (self User) Equal(other User) bool {
	if self.Id != other.Id {
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
