package user

import (
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/user"
)

func (s *impl) GetUsers() ([]user.User, error) {
	return s.userRepo.GetAll()
}
