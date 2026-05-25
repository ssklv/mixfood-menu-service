package handlers

import "io"

const (
	AccessCookie = "access_token"
)

type FileStorage interface {
	SaveFile(file io.Reader, filename string) (string, error)
}

type Logger interface {
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

type ErrorResponse struct {
	Error string `json:"error"`
}
