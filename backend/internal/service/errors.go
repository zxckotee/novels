package service

import "errors"

// Общие ошибки сервисов
var (
	ErrNovelNotFound = errors.New("novel not found")
	ErrNotFound      = errors.New("not found")
	ErrNotAuthorized = errors.New("not authorized")
	ErrInvalidAction = errors.New("invalid action")
)
