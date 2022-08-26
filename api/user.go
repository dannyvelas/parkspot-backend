package api

type user struct {
	Id           string `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Email        string `json:"email"`
	Role         role   `json:"role"`
	TokenVersion int    `json:"tokenVersion"`
}

func newUser(id string, firstName string, lastName string, email string, role role) user {
	return user{
		Id:        id,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Role:      role,
	}
}

func (self user) Equal(other user) bool {
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
	}
	return true
}
