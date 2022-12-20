package models

type CreateResident struct {
	ResidentID string `json:"residentID"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	UnlimDays  bool   `json:"unlimDays"`
}
