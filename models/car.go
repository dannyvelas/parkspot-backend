package models

type Car struct {
	Id           string `json:"id"`
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func NewCar(id string, licensePlate string, color string, make string, model string) Car {
	return Car{
		Id:           id,
		LicensePlate: licensePlate,
		Color:        color,
		Make:         make,
		Model:        model,
	}
}

func (self Car) Equal(other Car) bool {
	if self.Id != other.Id {
		return false
	} else if self.LicensePlate != other.LicensePlate {
		return false
	} else if self.Color != other.Color {
		return false
	} else if self.Make != other.Make {
		return false
	} else if self.Model != other.Model {
		return false
	}

	return true
}
