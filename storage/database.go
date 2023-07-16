package storage

type Database interface {
	AdminRepo() AdminRepo
	ResidentRepo() ResidentRepo
	CarRepo() CarRepo
	PermitRepo() PermitRepo
	VisitorRepo() VisitorRepo
}
