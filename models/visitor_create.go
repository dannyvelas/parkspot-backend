package models

type CreateVisitor struct {
	ResidentID   string
	FirstName    string
	LastName     string
	Relationship string
	StartTS      int64
	EndTS        int64
}
