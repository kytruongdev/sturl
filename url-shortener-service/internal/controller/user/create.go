package user

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/user"
)

func (s *impl) CreateUser(name, email string) (*user.User, error) {
	return s.userRepo.Create(name, email)
}
