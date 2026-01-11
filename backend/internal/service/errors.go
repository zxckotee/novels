package service

import "errors"

// Общие ошибки сервисов
var (
	ErrNovelNotFound = errors.New("novel not found")
)
