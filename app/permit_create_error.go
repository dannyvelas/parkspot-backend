package app

import (
	"fmt"
	"github.com/dannyvelas/lasvistas_api/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type CreatePermitError struct {
	err string
}

// implements error interface
func (e CreatePermitError) Error() string {
	return e.err
}

var (
	ErrResidentForPermitDNE = CreatePermitError{"Users must have a registered account to request" +
		" a guest parking permit. Please create their account before requesting their permit."}
	ErrCarForPermitDNE = CreatePermitError{"The car that you chose for this permit does not" +
		" exist. Please create or choose another car."}
	ErrCarActivePermit = CreatePermitError{"Cannot create a permit during these dates" +
		" because this car has at least one active permit during that time."}
	ErrPermitTooLong = CreatePermitError{
		err: fmt.Sprintf("Error: Requests cannot be longer than %d days,"+
			" unless there is an exception."+
			"\nIf this resident wants their guest to park for more than %d days, they"+
			" can apply for another request once that one expires.",
			config.MaxPermitLength,
			config.MaxPermitLength),
	}
	ErrResidentTwoActivePermits = CreatePermitError{"Cannot create a permit during these dates" +
		" because this resident has at least two active permits during that time."}
)

func errEntityDaysTooLong(entity string, amtDaysUsed int) CreatePermitError {
	entityLower := cases.Lower(language.English).String(entity)
	entityTitle := cases.Title(language.English).String(entity)

	return CreatePermitError{
		err: fmt.Sprintf("Error: This %s has used parking permits that have lasted"+
			" a combined total of %d days."+
			"\n%ss are allowed maximum %d days of parking passes, unless there is an exception."+
			"\nThis %s must wait until next year to give out new parking passes.",
			entityLower, amtDaysUsed,
			entityTitle, config.MaxParkingDays,
			entityLower),
	}
}

func errPermitPlusEntityDaysTooLong(entity string, amtDaysUsed int) CreatePermitError {
	entityLower := cases.Lower(language.English).String(entity)

	return CreatePermitError{
		err: fmt.Sprintf("Error: This request would exceed the %s's"+
			" yearly guest parking pass limit of %d days."+
			"\nThis %s has given out parking permits for a total of %d days."+
			"\nThis %s can give out max %d more day(s) before reaching their limit."+
			"\nThis %s can only give more permits if they have unlimited days or if"+
			" their requested permites are exceptions",
			entityLower, config.MaxParkingDays,
			entityLower, amtDaysUsed,
			entityLower, config.MaxParkingDays-amtDaysUsed,
			entityLower)}
}
