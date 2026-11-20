package storage

import "errors"

var (
	ErrQuotaExceeded    = errors.New("quota ledger: resident has already met or exceeded their effective allowance")
	ErrQuotaWouldExceed = errors.New("quota ledger: this permit would exceed resident's effective allowance")
)
