package models

type CreateResident struct {
	ResidentId string `json:"residentId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	UnlimDays  bool   `json:"unlimDays"`
}
