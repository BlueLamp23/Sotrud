package repository

// Кладовщик для таблицы appointments (назначения сотрудников на должности).
// Таблица appointments — связующая: сотрудники M:N должности.
// Один сотрудник может занимать много должностей ПОСЛЕДОВАТЕЛЬНО (история),
// но только ОДНУ одновременно (end_date IS NULL = текущая).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
)

// AppointmentRepo — кладовщик для appointments.
type AppointmentRepo struct {
	db *sql.DB
}

func NewAppointmentRepo(db *sql.DB) *AppointmentRepo {
	return &AppointmentRepo{db: db}
}

// Create добавляет назначение: сотрудник -> должность с даты startDate.
// end_date не передаём — только что созданное назначение всегда активно
// (end_date NULL в базе). Дата окончания появится при Close.
func (r *AppointmentRepo) Create(ctx context.Context, employeeID, positionID int64, startDate string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO appointments (employee_id, position_id, start_date) VALUES (?, ?, ?)",
		employeeID, positionID, startDate,
	)
	if err != nil {
		return 0, fmt.Errorf("create appointment: %w", err)
	}
	return res.LastInsertId()
}

// GetByID возвращает одно назначение по его id.
// Нужен в Close, чтобы перед UPDATE проверить: «а оно вообще активно?».
func (r *AppointmentRepo) GetByID(ctx context.Context, id int64) (*model.Appointment, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, employee_id, position_id, start_date, end_date FROM appointments WHERE id = ?",
		id,
	)

	var a model.Appointment
	var idValue, employeeID, positionID int64
	var startDate string
	var endDate sql.NullString
	err := row.Scan(&idValue, &employeeID, &positionID, &startDate, &endDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get appointment: %w", err)
	}
	a.ScanAppointment(idValue, employeeID, positionID, startDate, endDate)
	return &a, nil
}

// GetActiveByEmployee — ключевой метод для правила «одно активное назначение».
// Ищет назначение у сотрудника, у которого end_date IS NULL (не закрыто).
// Если такого нет — возвращает ErrNotFound (не ошибка).
//
// Сервис будет звать это ПЕРЕД Create:
func (r *AppointmentRepo) GetActiveByEmployee(ctx context.Context, employeeID int64) (*model.Appointment, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, employee_id, position_id, start_date, end_date "+
			"FROM appointments WHERE employee_id = ? AND end_date IS NULL",
		employeeID,
	)

	var a model.Appointment
	var idValue, rowEmployeeID, positionID int64
	var startDate string
	var endDate sql.NullString
	err := row.Scan(&idValue, &rowEmployeeID, &positionID, &startDate, &endDate)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get active appointment: %w", err)
	}
	a.ScanAppointment(idValue, rowEmployeeID, positionID, startDate, endDate)
	return &a, nil
}

// Close закрывает назначение: ставит end_date = endDate.
// Это «снять с должности». После этого сотрудник может быть
// назначен на другую должность.
//
// На вход идёт id САМОГО НАЗНАЧЕНИЯ (не сотрудника!), это важно.
// Если UI знает только сотрудника — сначала GetActiveByEmployee, потом Close.
func (r *AppointmentRepo) Close(ctx context.Context, id int64, endDate string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE appointments SET end_date = ? WHERE id = ? AND end_date IS NULL",
		endDate, id,
	)
	if err != nil {
		return fmt.Errorf("close appointment: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("close appointment rows: %w", err)
	}
	if n == 0 {
		// 0 строк — значит, либо такого id нет, либо уже закрыто.
		// В обоих случаях для сервиса это «не найдено активное».
		return ErrNotFound
	}
	return nil
}

// HistoryByEmployee — все назначения сотрудника с названием должности.
// JOIN с positions, чтобы не было повторного запроса на фронте.
// Сортируем: свежие сверху (ORDER BY start_date DESC).
func (r *AppointmentRepo) HistoryByEmployee(ctx context.Context, employeeID int64) ([]model.AppointmentWithPosition, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, p.title, a.start_date, a.end_date
		 FROM appointments a
		 INNER JOIN positions p ON p.id = a.position_id
		 WHERE a.employee_id = ?
		 ORDER BY a.start_date DESC`,
		employeeID,
	)
	if err != nil {
		return nil, fmt.Errorf("history appointments: %w", err)
	}
	defer rows.Close()

	var result []model.AppointmentWithPosition
	for rows.Next() {
		var a model.AppointmentWithPosition
		var id int64
		var title, startDate string
		var endDate sql.NullString
		if err := rows.Scan(&id, &title, &startDate, &endDate); err != nil {
			return nil, fmt.Errorf("scan appointment: %w", err)
		}
		a.ScanAppointmentWithPosition(id, title, startDate, endDate)
		result = append(result, a)
	}
	return result, rows.Err()
}
