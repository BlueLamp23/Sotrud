package service

//  Здесь ПРАВИЛА, не SQL.
//проверить → решить → вызвать репозиторий.

import (
	"context" // errors.New для объявления ошибок
	"errors"
	"strings" // strings.TrimSpace обрезать пробелы

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
)

// PositionService  -  «начальник по должностям».
// Внутри  -  репозиторий, через который сервис ходит в базу.
type PositionService struct {
	repo *repository.PositionRepo
}

// NewPositionService — конструктор. Принимает готовый репозиторий.
func NewPositionService(repo *repository.PositionRepo) *PositionService {
	return &PositionService{repo: repo}
}

// Create — добавить должность.
// Правила: title не пустой (после trim), длина ≤ 100 символов.
func (s *PositionService) Create(ctx context.Context, title string, isActive bool) (int64, error) {
	// TrimSpace убирает пробелы по краям: "  Инженер  " → "Инженер".
	// Без этого "   " пройдёт проверку "не пустой" и запишется в базу.
	title = strings.TrimSpace(title)

	if title == "" {
		return 0, ErrValidation
	}
	if len(title) > 100 {
		return 0, ErrValidation
	}

	return s.repo.Create(ctx, title, isActive)
}

// List — список всех должностей. Тут правил нет, просто проброс.
// Даже такой метод пишем: handler не должен знать про репозиторий.
func (s *PositionService) List(ctx context.Context) ([]model.Position, error) {
	return s.repo.List(ctx)
}

// Update — изменить должность.
// Те же правила, что в Create, плюс: должность должна существовать
// (это проверяет сам repo.Update через RowsAffected, вернёт ErrNotFound).
func (s *PositionService) Update(ctx context.Context, id int64, title string, isActive bool) error {
	title = strings.TrimSpace(title)

	if title == "" {
		return ErrValidation
	}
	if len(title) > 100 {
		return ErrValidation
	}

	return s.repo.Update(ctx, id, title, isActive)
}

// Delete — удалить должность.
// ГЛАВНОЕ ПРАВИЛО: нельзя удалить должность, на которой кто-то работает сейчас.
//
// Пошагово:
//  1. Спросить у репозитория: сколько АКТИВНЫХ назначений на эту должность?
//  2. Если > 0 → вернуть ErrPositionInUse (handler отдаст 409 Conflict).
//  3. Если 0 → удалить через репозиторий.
//
// Почему не полагаемся только на ON DELETE RESTRICT из init.sql?
// Потому что тогда пользователь увидит "FOREIGN KEY constraint failed" —
// техническую ошибку на английском. А мы хотим "на должности есть сотрудники".
func (s *PositionService) Delete(ctx context.Context, id int64) error {
	// Сначала проверяем активные — это быстро и даёт красивое сообщение.
	n, err := s.repo.CountActiveAppointments(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return ErrPositionInUse
	}

	// Если активных нет, но закрытые есть — repo.Delete вернёт ErrFKViolation.
	// Это тоже «нельзя удалить», просто по другой причине.
	err = s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrFKViolation) {
		return ErrPositionInUse
	}
	return err
}
