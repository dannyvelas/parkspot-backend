package storage

type Database interface {
	Ping() error
	AdminRepo() AdminRepo
	ResidentRepo() ResidentRepo
	CarRepo() CarRepo
	PermitRepo() PermitRepo
	VisitorRepo() VisitorRepo
	QuotaLedgerRepo() QuotaLedgerRepo
}
