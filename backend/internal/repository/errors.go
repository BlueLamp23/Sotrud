package repository

import "errors"

// ErrNotFound — «запись не найдена». Используется всеми репозиториями.
var ErrNotFound = errors.New("not found")

// ErrFKViolation — нарушение внешнего ключа (ON DELETE RESTRICT).
// Возникает, когда удаляют должность, на которой есть назначения.
var ErrFKViolation = errors.New("foreign key violation")
