package repository

// Пакет repository - слой доступа к данным.

import (
	"context"      // первый параметр каждого метода
	"database/sql" // интерфейс базы и sql.ErrNoRows
	"errors"
	"fmt" // fmt.Errorf
	"strings"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
)

// ErrNotFound - общая ошибка «запись не найдена».
//
//	один раз для всех репозиториев.

// Хранит внутри указатель на базу — через него делает все запросы.
type PositionRepo struct {
	db *sql.DB
}

// NewPositionRepo - конструктор. Принимает уже открытую базу (её открывает NewDB).
func NewPositionRepo(db *sql.DB) *PositionRepo {
	return &PositionRepo{db: db}
}

// Create вставляет должность и возвращает её новый id.
// isActive передаём как bool — в SQLite переводим в 0/1.
func (r *PositionRepo) Create(ctx context.Context, title string, isActive bool) (int64, error) {
	// bool -> int 0/1: в SQLite bool не хранится, только INTEGER.
	active := 0
	if isActive {
		active = 1
	}

	// служебная инфа(id последней вставки, число строк).
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO positions (title, is_active) VALUES (?, ?)",
		title, active,
	)
	if err != nil {
		// %w - обернуть ошибку, сохранив её для errors.is
		return 0, fmt.Errorf("create position: %w", err)
	}

	// LastInsertId - id только что вставленной строки.
	return res.LastInsertId()
}

// GetByID возвращает одну должность по id.
// Если строки нет - возвращает ErrNotFound.
func (r *PositionRepo) GetByID(ctx context.Context, id int64) (*model.Position, error) {
	// QueryRowContext запрос, который вернёт максимум одну строку.
	row := r.db.QueryRowContext(ctx,
		"SELECT id, title, is_active FROM positions WHERE id = ?", id)

	var p model.Position
	var active int // сюда читаем 0/1 из базы
	err := row.Scan(&p.ID, &p.Title, &active)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound // ноль строк -  404
	}
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}

	// int -> bool: 1 = true, всё остальное = false.
	p.IsActive = active == 1
	return &p, nil
}

// List возвращает все должности, отсортированные по названию.
// Для списка в интерфейсе.
func (r *PositionRepo) List(ctx context.Context) ([]model.Position, error) {
	// QueryContext — запрос, возвращающий много строк.
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, is_active FROM positions ORDER BY title")
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}
	defer rows.Close() // обязательно: освободить соединение, когда прочитали всё

	var result []model.Position
	for rows.Next() { // Next() возвращает true, пока есть ещё строки
		var p model.Position
		var active int
		if err := rows.Scan(&p.ID, &p.Title, &active); err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		p.IsActive = active == 1
		result = append(result, p)
	}
	// rows.Err() проверка, не было ли ошибки во время итерации.
	return result, rows.Err()
}

// Update перезаписывает название и активность.
// Если строки с таким id нет  возвращает ErrNotFound.
func (r *PositionRepo) Update(ctx context.Context, id int64, title string, isActive bool) error {
	active := 0
	if isActive {
		active = 1
	}

	res, err := r.db.ExecContext(ctx,
		"UPDATE positions SET title = ?, is_active = ? WHERE id = ?",
		title, active, id,
	)
	if err != nil {
		return fmt.Errorf("update position: %w", err)
	}

	// RowsAffected — сколько строк реально изменилось.
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update position rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound // UPDATE не задел ни одной строки = такого id нет
	}
	return nil
}

// Delete удаляет должность по id.
// если на должности есть сотрудники, SQLite вернёт ошибку
// FOREIGN KEY constraint failed - в init.sql стоит ON DELETE RESTRICT.
func (r *PositionRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM positions WHERE id = ?", id)
	if err != nil {
		if strings.Contains(err.Error(), "FOREIGN KEY") {
			return ErrFKViolation
		}
		return fmt.Errorf("delete position: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete position rows: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CountActiveAppointments - сколько сотрудников СЕЙЧАС на этой должности.
// end_date IS NULL = назначение ещё не закрыто, значит активно.
// Нужен сервису для правила «нельзя удалить занятую должность».
func (r *PositionRepo) CountActiveAppointments(ctx context.Context, positionID int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM appointments WHERE position_id = ? AND end_date IS NULL",
		positionID,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count active appointments: %w", err)
	}
	return n, nil
}
