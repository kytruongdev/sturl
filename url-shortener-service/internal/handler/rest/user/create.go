package user

import (
	"encoding/json"
	"net/http"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/httpserver"
)

func (h *Handler) CreateUser() http.HandlerFunc {
	return httpserver.HandlerErr(func(w http.ResponseWriter, r *http.Request) error {
		var payload struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {

			return httpserver.Error{
				Status: http.StatusBadRequest,
				Code:   "invalid_request",
				Desc:   "Invalid request",
			}
		}

		user, err := h.userCtrl.CreateUser(payload.Name, payload.Email)
		if err != nil {
			return httpserver.Error{
				Status: http.StatusInternalServerError,
				Code:   "failed_to_create_user",
				Desc:   "Failed to create user",
			}
		}

		httpserver.RespondJSON(w, user)

		return nil
	})
}
