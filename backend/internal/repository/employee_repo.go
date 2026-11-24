package repository

// Кладовщик для таблицы employees + JOIN с appointments и positions,
// чтобы главный экран сразу показывал текущую должность.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
)

// EmployeeRepo — кладовщик для employees.
type EmployeeRepo struct {
	db *sql.DB
}

func NewEmployeeRepo(db *sql.DB) *EmployeeRepo {
	return &EmployeeRepo{db: db}
}

// optionalString превращает string в sql.NullString:
// пустая строка -> NULL (в базу), непустая -> значение.
// Нужно, потому что phone и email в init.sql — nullable.
func optionalString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false} // будет записано как NULL
	}
	return sql.NullString{String: s, Valid: true}
}

// Create вставляет сотрудника и возвращает id.
func (r *EmployeeRepo) Create(ctx context.Context, fullName, phone, email, status string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO employees (full_name, phone, email, status) VALUES (?, ?, ?, ?)",
		fullName, optionalString(phone), optionalString(email), status,
	)
	if err != nil {
		return 0, fmt.Errorf("create employee: %w", err)
	}
	return res.LastInsertId()
}

// GetByID возвращает одного сотрудника (без должности — она в List).
func (r *EmployeeRepo) GetByID(ctx context.Context, id int64) (*model.Employee, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, full_name, phone, email, status FROM employees WHERE id = ?", id)

	var e model.Employee
	var rowID int64
	var fullName, status string
	var phone, email sql.NullString
	err := row.Scan(&rowID, &fullName, &phone, &email, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get employee: %w", err)
	}
	e.ScanEmployee(rowID, fullName, phone, email, status)
	return &e, nil
}

// List — главный метод: список сотрудников с текущей должностью.
// Параметры фильтрации:
//
//	search     — подстрока ФИО (если пусто — не фильтруем)
//	status     — 'Работает' / 'Уволен' (если пусто — все)
//	positionID — id должности (0 — не фильтруем)
//
// Логика SQL:
//
//	FROM employees e
//	LEFT JOIN appointments a  — берём назначения сотрудника
//	    ON a.employee_id = e.id AND a.end_date IS NULL  <- только АКТИВНОЕ
//	LEFT JOIN positions p     — название должности
//	    ON p.id = a.position_id
//
// LEFT JOIN, а не INNER — чтобы сотрудник БЕЗ назначения тоже попал в список.
func (r *EmployeeRepo) List(ctx context.Context, search, status string, positionID int64) ([]model.EmployeeWithPosition, error) {
	// Собираем WHERE динамически: если search пустой — не добавляем условие.
	// placeholder'ы ? — чтобы не было SQL-инъекции.
	query := `
		SELECT
			e.id,
			e.full_name,
			e.status,
			p.title,
			a.start_date
		FROM employees e
		LEFT JOIN appointments a
			ON a.employee_id = e.id AND a.end_date IS NULL
		LEFT JOIN positions p
			ON p.id = a.position_id
		WHERE 1=1`
	args := []any{}

	if search != "" {
		query += " AND e.full_name LIKE ?"
		args = append(args, "%"+search+"%") // %...% = "содержит подстроку"
	}
	if status != "" {
		query += " AND e.status = ?"
		args = append(args, status)
	}
	// Фильтр по должности: сотрудник должен занимать её СЕЙЧАС (активное назначение).
	if positionID > 0 {
		query += " AND a.position_id = ?"
		args = append(args, positionID)
	}
	query += " ORDER BY e.full_name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list employees: %w", err)
	}
	defer rows.Close()

	var result []model.EmployeeWithPosition
	for rows.Next() {
		var e model.EmployeeWithPosition
		var employeeID int64
		var fullName, empStatus string
		var title, startDate sql.NullString
		if err := rows.Scan(&employeeID, &fullName, &empStatus, &title, &startDate); err != nil {
			return nil, fmt.Errorf("scan employee: %w", err)
		}
		e.ScanEmployeeWithPosition(employeeID, fullName, empStatus, title, startDate)
		result = append(result, e)
	}
	return result, rows.Err()
}

// Update перезаписывает поля сотрудника.
func (r *EmployeeRepo) Update(ctx context.Context, id int64, fullName, phone, email, status string) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE employees SET full_name = ?, phone = ?, email = ?, status = ? WHERE id = ?",
		fullName, optionalString(phone), optionalString(email), status, id,
	)
	if err != nil {
		return fmt.Errorf("update employee: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update employee rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete удаляет сотрудника.
// В init.sql на appointments.employee_id стоит ON DELETE CASCADE —
// значит, вместе с сотрудником удалятся все его назначения. Это по ТЗ.
func (r *EmployeeRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM employees WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete employee: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete employee rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ExistsByEmail — есть ли уже сотрудник с таким email.
// Нужно сервису: «email должен быть уникальным».
// Пустой email не считаем (у нескольких сотрудников может не быть email).
func (r *EmployeeRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, nil
	}
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM employees WHERE email = ?", email,
	).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("exists by email: %w", err)
	}
	return n > 0, nil
}
