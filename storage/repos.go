package storage

type Repos struct {
	Admin    AdminRepo
	Permit   PermitRepo
	Car      CarRepo
	Resident ResidentRepo
	Visitor  VisitorRepo
}

func NewRepos(database Database) Repos {
	return Repos{
		Admin:    NewAdminRepo(database),
		Permit:   NewPermitRepo(database),
		Car:      NewCarRepo(database),
		Resident: NewResidentRepo(database),
		Visitor:  NewVisitorRepo(database),
	}
}
