package httpserver

// Response represents a structured HTTP response with status code and data payload.
type Response struct {
	Status int         `json:"status"` // HTTP status code
	Data   interface{} `json:"data"`   // Response data payload
}
