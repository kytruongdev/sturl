package user

import (
	"context"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/models"
)

func (r *impl) GetAll() ([]User, error) {
	users, err := models.Users().All(context.Background(), r.db)
	if err != nil {
		return nil, err
	}

	var result []User
	for _, u := range users {
		result = append(result, User{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}
	return result, nil
}
