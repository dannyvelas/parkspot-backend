package models

type Admin struct {
	Id           string
	FirstName    string
	LastName     string
	Email        string
	Password     string
	IsPrivileged bool
}
