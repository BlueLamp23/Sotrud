package handler

// Помощники для всех handler'ов:
//   writeJSON  — отдать JSON с кодом
//   writeError  — отдать {"error": "..."} с кодом
//   writeServiceError — перевести ошибку сервиса в HTTP-код

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// writeJSON — единая точка для ответов с данными.
// Ставит Content-Type, код, кодирует data в JSON.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("writeJSON: encode error: %v", err)
		}
	}
}

// writeError — единая точка для ошибок.
// Формат: {"error": "сообщение"}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeServiceError берёт ошибку из service/repository и решает какой HTTP-код отдать.
//
// Таблица:
//
//	service.ErrValidation         → 400  (пользователь ввёл не то)
//	service.ErrInvalidStatus      → 400
//	service.ErrInvalidCredentials → 401  (неверный логин/пароль/токен)
//	repository.ErrNotFound        → 404  (записи нет)
//	service.ErrPositionInUse      → 409  (конфликт: нельзя удалить занятую)
//	service.ErrEmailTaken         → 409  (конфликт: email занят)
//	service.ErrAlreadyAssigned    → 409  (конфликт: уже назначен)
//	service.ErrLoginTaken         → 409  (конфликт: логин занят)
//	всё остальное                 → 500  (баг сервера)
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrValidation):
		writeError(w, http.StatusBadRequest, "validation error")

	case errors.Is(err, service.ErrInvalidStatus):
		writeError(w, http.StatusBadRequest, "invalid status")

	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")

	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")

	case errors.Is(err, service.ErrPositionInUse):
		writeError(w, http.StatusConflict, "position has active appointments")

	case errors.Is(err, service.ErrEmailTaken):
		writeError(w, http.StatusConflict, "email already taken")

	case errors.Is(err, service.ErrAlreadyAssigned):
		writeError(w, http.StatusConflict, "employee already has an active appointment")

	case errors.Is(err, service.ErrLoginTaken):
		writeError(w, http.StatusConflict, "login already taken")

	default:
		log.Printf("unhandled error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
