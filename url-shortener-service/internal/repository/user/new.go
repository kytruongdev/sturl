package user

import "database/sql"

type User struct {
	ID    int
	Name  string
	Email string
}

type Repository interface {
	GetAll() ([]User, error)
	Create(name, email string) (*User, error)
}

type impl struct {
	db *sql.DB
}

func New(db *sql.DB) Repository {
	return &impl{db: db}
}
