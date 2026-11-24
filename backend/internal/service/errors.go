package service

// Все ошибки уровня «правило нарушено» — в одном месте.
// Handler превращает их в HTTP-коды:
//   ErrValidation         → 400 Bad Request
//   ErrPositionInUse      → 409 Conflict
//   ErrEmailTaken         → 409 Conflict
//   ErrAlreadyAssigned    → 409 Conflict
//   ErrLoginTaken         → 409 Conflict
//   ErrInvalidCredentials → 401 Unauthorized
//   ErrInvalidStatus      → 400 Bad Request
//   repository.ErrNotFound → 404 Not Found (объявлена в repository, не здесь)

import "errors" //ну понятно

var (
	// ErrValidation — общая ошибка «пользователь ввёл что-то не то».
	// Используется во всех сервисах, поэтому объявлена здесь, а не в каждом файле.
	ErrValidation = errors.New("validation error")

	// Должности
	ErrPositionInUse = errors.New("position has active appointments")

	// Сотрудники
	ErrEmailTaken    = errors.New("email already taken")
	ErrInvalidStatus = errors.New("invalid status")

	// Назначения
	ErrAlreadyAssigned = errors.New("employee already has an active appointment")

	// Аутентификация
	ErrLoginTaken         = errors.New("login already taken")
	ErrInvalidCredentials = errors.New("invalid login or password")
)
