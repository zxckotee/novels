package response

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response представляет стандартный ответ API
type Response struct {
	Data interface{} `json:"data,omitempty"`
	Meta *Meta       `json:"meta,omitempty"`
}

// Meta содержит метаданные ответа
type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error *ErrorDetail `json:"error"`
	Meta  *Meta        `json:"meta,omitempty"`
}

// ErrorDetail содержит детали ошибки
type ErrorDetail struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldError  `json:"details,omitempty"`
}

// FieldError представляет ошибку валидации поля
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// PaginatedResponse представляет ответ с пагинацией
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
	Meta       *Meta       `json:"meta,omitempty"`
}

// Pagination содержит информацию о пагинации
type Pagination struct {
	CurrentPage int  `json:"current_page"`
	PerPage     int  `json:"per_page"`
	TotalPages  int  `json:"total_pages"`
	TotalItems  int  `json:"total_items"`
	HasNext     bool `json:"has_next"`
	HasPrev     bool `json:"has_prev"`
}

// NewPagination создает новый объект пагинации
func NewPagination(page, perPage, totalItems int) *Pagination {
	totalPages := totalItems / perPage
	if totalItems%perPage > 0 {
		totalPages++
	}

	return &Pagination{
		CurrentPage: page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}
}

// JSON отправляет JSON ответ
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	resp := Response{
		Data: data,
		Meta: &Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// Paginated отправляет ответ с пагинацией
func Paginated(w http.ResponseWriter, statusCode int, data interface{}, pagination *Pagination) {
	resp := PaginatedResponse{
		Data:       data,
		Pagination: pagination,
		Meta: &Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// Error отправляет ответ с ошибкой
func Error(w http.ResponseWriter, statusCode int, code, message string) {
	resp := ErrorResponse{
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
		},
		Meta: &Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// ValidationError отправляет ответ с ошибками валидации
func ValidationError(w http.ResponseWriter, errors []FieldError) {
	resp := ErrorResponse{
		Error: &ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: errors,
		},
		Meta: &Meta{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	json.NewEncoder(w).Encode(resp)
}

// Created отправляет ответ 201 Created
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// OK отправляет ответ 200 OK
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

// NoContent отправляет ответ 204 No Content
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// NotFound отправляет ответ 404 Not Found
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// BadRequest отправляет ответ 400 Bad Request
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// InternalError отправляет ответ 500 Internal Server Error
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

// Unauthorized отправляет ответ 401 Unauthorized
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden отправляет ответ 403 Forbidden
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// Conflict отправляет ответ 409 Conflict
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}
