package service

// У сотрудника не может быть двух активных назначений одновременно
// В init.sql этого правила нет
import (
	"context"
	"strings"
	"time"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
)

// dateLayout — формат дат в базе
const dateLayout = "2006-01-02"

// Держит три репозитория: свой + два для проверок существования
type AppointmentService struct {
	appointmentRepo *repository.AppointmentRepo
	employeeRepo    *repository.EmployeeRepo
	positionRepo    *repository.PositionRepo
}

// NewAppointmentService — конструктор
func NewAppointmentService(
	appointmentRepo *repository.AppointmentRepo,
	employeeRepo *repository.EmployeeRepo,
	positionRepo *repository.PositionRepo,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		employeeRepo:    employeeRepo,
		positionRepo:    positionRepo,
	}
}

// разобрать строку в time.Time.
// Если формат неверный возвращает ошибку, которую caller превратит в ErrValidation.
func parseDate(s string) (time.Time, error) {
	return time.Parse(dateLayout, s)
}

// Assign — назначить сотрудника на должность.
//  1. Нормализовать дату (trim), проверить формат YYYY-MM-DD.
//  2. Проверить, что сотрудник существует (иначе → ErrNotFound → 404).
//  3. Проверить, что должность существует и активна.
//  4. Проверить, что у сотрудника нет активного назначения.
//  5. Создать назначение.
func (s *AppointmentService) Assign(ctx context.Context, employeeID, positionID int64, startDate string) (int64, error) {
	// 1. Дата.
	startDate = strings.TrimSpace(startDate)
	if _, err := parseDate(startDate); err != nil {
		return 0, ErrValidation
	}

	// 2. Сотрудник существует?
	if _, err := s.employeeRepo.GetByID(ctx, employeeID); err != nil {
		return 0, err // сюда попадёт repository.ErrNotFound → handler отдаст 404
	}

	// 3. Должность существует и активна?
	// Неактивную должность назначать нельзя: она убрана из справочника (is_active = 0).
	position, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		return 0, err
	}
	if !position.IsActive {
		return 0, ErrValidation
	}

	// 4. одно активное назначение.
	// err == nil → активное уже есть → ErrAlreadyAssigned
	// err == ErrNotFound  → активного нет, идём дальше
	_, err = s.appointmentRepo.GetActiveByEmployee(ctx, employeeID)
	if err == nil {
		return 0, ErrAlreadyAssigned
	}
	if err != repository.ErrNotFound {
		return 0, err
	}

	// 5. Создать.
	return s.appointmentRepo.Create(ctx, employeeID, positionID, startDate)
}

// Unassign — снять сотрудника с должности.
//  1. Нормализовать дату (trim), проверить формат.
//  2. Найти активное назначение сотрудника. Если его нет → ErrNotFound → 404
//  3. Проверить, что endDate >= startDate найденного назначения.
//     (Иначе нарушится CHECK из init.sql, и SQLite вернёт непонятную ошибку.)
//  4. Закрыть назначение.
func (s *AppointmentService) Unassign(ctx context.Context, employeeID int64, endDate string) error {
	// 1. Дата.
	endDate = strings.TrimSpace(endDate)
	endParsed, err := parseDate(endDate)
	if err != nil {
		return ErrValidation
	}

	// 2. Найти активное.
	active, err := s.appointmentRepo.GetActiveByEmployee(ctx, employeeID)
	if err != nil {
		// ErrNotFound handler отдаст 404.
		// Другую ошибку handler отдаст 500.
		return err
	}

	// 3. endDate не раньше startDate.
	startParsed, err := parseDate(active.StartDate)
	if err != nil {
		// В базе лежит мусор вместо даты, но нам надо как-то ответить. Возвращаем ErrValidation,
		return ErrValidation
	}
	if endParsed.Before(startParsed) {
		return ErrValidation // дата окончания раньше даты начала
	}

	// 4. Закрыть. Передаём id назначения, не сотрудника.
	return s.appointmentRepo.Close(ctx, active.ID, endDate)
}

// History — история назначений сотрудника.
// Правил нет: проброс в репозиторий. Если сотрудник не существует — вернётся пустой список - не ошибка
func (s *AppointmentService) History(ctx context.Context, employeeID int64) ([]model.AppointmentWithPosition, error) {
	return s.appointmentRepo.HistoryByEmployee(ctx, employeeID)
}
