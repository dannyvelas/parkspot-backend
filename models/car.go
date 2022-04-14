package models

type Car struct {
	Id           string `json:"id"`
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}
