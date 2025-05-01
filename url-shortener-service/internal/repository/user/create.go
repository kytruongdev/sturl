package user

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/models"
	"github.com/volatiletech/sqlboiler/boil"
)

func (r *impl) Create(name, email string) (*User, error) {
	u := models.User{
		Name:  name,
		Email: email,
	}

	err := u.Insert(context.Background(), r.db, boil.Infer())
	if err != nil {
		return nil, err
	}

	return &User{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}, nil
}
