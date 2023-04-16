package validator

import (
	"github.com/dannyvelas/lasvistas_api/errs"
)

func validateEditAmtDays(amtDays *int) error {
	if amtDays == nil {
		return nil
	}

	if *amtDays < 0 {
		return errs.InvalidFields("amtDays cannot be lower than 0")
	}

	return nil
}
