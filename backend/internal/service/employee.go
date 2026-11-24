package service

// Правила для сотрудников. SQL — в employee_repo.go, здесь — проверки.
import (
	"context"
	"strings"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
)

// Допустимые значения статуса в CHECK из init.sql.
// Вынесены в константы, чтобы не опечататься в строках.
const (
	StatusWorking = "Работает"
	StatusFired   = "Уволен"
)

// EmployeeService — «начальник по сотрудникам»
type EmployeeService struct {
	repo *repository.EmployeeRepo
}

// NewEmployeeService — конструктор
func NewEmployeeService(repo *repository.EmployeeRepo) *EmployeeService {
	return &EmployeeService{repo: repo}
}

// validate — ТОЛЬКО проверка, без нормализации.
func (s *EmployeeService) validate(ctx context.Context, fullName, email, status string) error {
	if fullName == "" {
		return ErrValidation
	}
	if len(fullName) > 200 {
		return ErrValidation
	}

	if status != StatusWorking && status != StatusFired {
		return ErrInvalidStatus
	}

	// Email: если задан — должен содержать @.
	if email != "" && !strings.Contains(email, "@") {
		return ErrValidation
	}

	return nil
}

// Create — добавить сотрудника.
// Порядок: нормализовать → проверить → проверить уникальность → вставить.
func (s *EmployeeService) Create(ctx context.Context, fullName, phone, email, status string) (int64, error) {
	// 1. Нормализация: убираем пробелы по краям.
	fullName = strings.TrimSpace(fullName)
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	status = strings.TrimSpace(status)

	// 2. Проверки формата.
	if err := s.validate(ctx, fullName, email, status); err != nil {
		return 0, err
	}

	// 3. Уникальность email (только если задан).
	if email != "" {
		exists, err := s.repo.ExistsByEmail(ctx, email)
		if err != nil {
			return 0, err
		}
		if exists {
			return 0, ErrEmailTaken
		}
	}

	// 4. Вставка. Сюда уходят уже обрезанные значения.
	return s.repo.Create(ctx, fullName, phone, email, status)
}

// List — список сотрудников с текущей должностью.
// Правил нет: search/status/positionID — просто параметры фильтра.
// Если status задан, но невалидный — отсеиваем.
func (s *EmployeeService) List(ctx context.Context, search, status string, positionID int64) ([]model.EmployeeWithPosition, error) {
	search = strings.TrimSpace(search)
	status = strings.TrimSpace(status)

	if status != "" && status != StatusWorking && status != StatusFired {
		return nil, ErrInvalidStatus
	}

	return s.repo.List(ctx, search, status, positionID)
}

// Get — один сотрудник по id
func (s *EmployeeService) Get(ctx context.Context, id int64) (*model.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

// Update — изменить сотрудника.
// email может остаться прежним.
func (s *EmployeeService) Update(ctx context.Context, id int64, fullName, phone, email, status string) error {
	// 1. Нормализация.
	fullName = strings.TrimSpace(fullName)
	email = strings.TrimSpace(email)
	phone = strings.TrimSpace(phone)
	status = strings.TrimSpace(status)

	// 2. Проверки формата.
	if err := s.validate(ctx, fullName, email, status); err != nil {
		return err
	}

	// 3. Уникальность email — только если он ИЗМЕНИЛСЯ.
	// Если тот же email, что у этого же сотрудника, — это не нарушение.
	if email != "" {
		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			// Сюда попадёт repository.ErrNotFound, если id не существует —
			// пробрасываем наверх, handler отдаст 404.
			return err
		}

		// Email изменился — проверяем, не занят ли он другим.
		// current.Email == nil значит, что у сотрудника email не был заполнен.
		if current.Email == nil || *current.Email != email {
			exists, err := s.repo.ExistsByEmail(ctx, email)
			if err != nil {
				return err
			}
			if exists {
				return ErrEmailTaken
			}
		}
	}

	// 4. Обновление.
	return s.repo.Update(ctx, id, fullName, phone, email, status)
}

// Delete — удалить сотрудника.
// Правил нет: ON DELETE CASCADE в init.sql удалит его назначения.
func (s *EmployeeService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
