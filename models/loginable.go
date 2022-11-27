package models

type Loginable interface {
	GetPassword() string
	AsUser() User
}
