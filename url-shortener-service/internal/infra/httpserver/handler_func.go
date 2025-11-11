package httpserver

import "net/http"

// HandlerErr wraps an error-returning handler function into a standard http.HandlerFunc.
// It converts handler functions that return errors into http.HandlerFunc by automatically
// handling errors and sending appropriate JSON responses.
func HandlerErr(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := h(writer, request); err != nil {
			RespondJSON(writer, err)
		}
	}
}
