package models

type Car struct {
	Id                 string `json:"id"`
	LicensePlate       string `json:"licensePlate"`
	Color              string `json:"color"`
	Make               string `json:"make"`
	Model              string `json:"model"`
	AmtParkingDaysUsed int    `json:"amtParkingDaysUsed"`
}

func NewCar(id string, licensePlate string, color string, make string, model string, amtParkingDaysUsed int) Car {
	return Car{
		Id:                 id,
		LicensePlate:       licensePlate,
		Color:              color,
		Make:               make,
		Model:              model,
		AmtParkingDaysUsed: amtParkingDaysUsed,
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
	} else if self.AmtParkingDaysUsed != other.AmtParkingDaysUsed {
		return false
	}

	return true
}

type NewCarArgs struct {
	LicensePlate string `json:"licensePlate"`
	Color        string `json:"color"`
	Make         string `json:"make"`
	Model        string `json:"model"`
}

func NewNewCarArgs(licensePlate string, color string, make string, model string) NewCarArgs {
	return NewCarArgs{
		LicensePlate: licensePlate,
		Color:        color,
		Make:         make,
		Model:        model,
	}
}

func (newCarArgs NewCarArgs) ToCar(id string) Car {
	return NewCar(id, newCarArgs.LicensePlate, newCarArgs.Color, newCarArgs.Make, newCarArgs.Model, 0)
}
