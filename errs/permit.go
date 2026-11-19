package errs

import "errors"

var ErrResidentQuotaFieldsNil = errors.New("resident's UnlimDays or AmtParkingDaysUsed is nil")
