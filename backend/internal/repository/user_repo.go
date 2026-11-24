package repository

// Кладовщик для таблицы users.
// Пользователь  это администратор, тот кто входит в систему.
import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
)

// UserRepo — кладовщик для users.
type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create вставляет нового пользователя и возвращает id.
// passwordHash — уже готовый хеш (его считает service/auth.go).
// Репозиторий НЕ хеширует
func (r *UserRepo) Create(ctx context.Context, login, passwordHash, role string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users (login, password_hash, role) VALUES (?, ?, ?)",
		login, passwordHash, role,
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return res.LastInsertId()
}

// GetByLogin возвращает пользователя по логину.
// Это вход в систему: сервис возьмёт PasswordHash из результата и сравнит.
// Возвращаем указатель на model.User (не на отдельные поля),
// чтобы сервис не мог перепутать порядок: он получает всю структуру.
func (r *UserRepo) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, login, password_hash, role FROM users WHERE login = ?",
		login,
	)

	var u model.User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by login: %w", err)
	}
	return &u, nil
}

// ExistsByLogin есть ли уже пользователь с таким логином.
// Нужно при регистрации: «логин занят»  409 Conflict.
// Возвращает bool, а не ErrNotFound потому что «нет такого» здесь, это не ошибка
func (r *UserRepo) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE login = ?", login,
	).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("exists by login: %w", err)
	}
	return n > 0, nil
}
