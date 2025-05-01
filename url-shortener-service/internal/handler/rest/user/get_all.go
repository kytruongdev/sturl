package user

import (
	"net/http"
	
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

func (h *Handler) GetUsers() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		users, err := h.userCtrl.GetUsers()
		if err != nil {
			return httpserver.Error{
				Status: http.StatusInternalServerError,
				Code:   "failed_to_get_users",
				Desc:   "Failed to get users",
			}
		}

		httpserver.RespondJSON(w, users)

		return nil
	})
}
