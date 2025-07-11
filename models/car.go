package models

type Car struct {
	ID                 string `json:"id"`
	ResidentID         string `json:"residentID"`
	LicensePlate       string `json:"licensePlate"`
	Color              string `json:"color"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	AmtParkingDaysUsed *int   `json:"amtParkingDaysUsed"`
}

func NewCar(id, residentID, licensePlate, color, make, model string, amtParkingDaysUsed int) Car {
	return Car{
		ID:                 id,
		ResidentID:         residentID,
		LicensePlate:       licensePlate,
		Color:              color,
		Make:               make,
		Model:              model,
		AmtParkingDaysUsed: &amtParkingDaysUsed,
	}
}

func (self Car) Equal(other Car) bool {
	if self.ID != other.ID {
		return false
	} else if self.ResidentID != other.ResidentID {
		return false
	} else if self.LicensePlate != other.LicensePlate {
		return false
	} else if self.Color != other.Color {
		return false
	} else if self.Make != other.Make {
		return false
	} else if self.Model != other.Model {
		return false
	} else if self.AmtParkingDaysUsed != other.AmtParkingDaysUsed {
		return false
	}

	return true
}
