package service

// Аутентификация: вход, выход, проверка токена, создание пользователей админом.
import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepo

	// mu защищает tokens от гонок: HTTP-запросы идут параллельно,
	mu     sync.RWMutex
	tokens map[string]*model.User
}

// NewAuthService — конструктор
func NewAuthService(userRepo *repository.UserRepo) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		tokens:   make(map[string]*model.User),
	}
}

// hashPassword — SHA-256 от пароля, в hex
// sha256.Sum256 возвращает [32]byte, hex.EncodeToString делает из неё строку
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

// generateToken — случайная 32-байтная строка в hex (64 символа)
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Login — вход. Проверяет пароль и выдаёт токен.
//
// Принцип:
//   - находим пользователя по логину;
//   - хешируем введённый пароль SHA-256;
//   - сравниваем с password_hash из базы.
//
// Ошибка ErrInvalidCredentials одинаковая для «нет такого логина» и «неверный пароль». Это скрывает, какие логины существуют.
func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	login = strings.TrimSpace(login)

	if login == "" || password == "" {
		return "", ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Единственная проверка: sha256(введённый) == password_hash из базы.
	if user.PasswordHash != hashPassword(password) {
		return "", ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.tokens[token] = user
	s.mu.Unlock()

	return token, nil
}

// Logout — «выйти». Удаляет токен из памяти.
// Если токена не было не ошибка (delete безопасен для отсутствующего ключа).
func (s *AuthService) Logout(token string) {
	s.mu.Lock()
	delete(s.tokens, token)
	s.mu.Unlock()
}

// UserByToken найти пользователя по токену.
// Используется в middleware для проверки каждого защищённого запроса.
func (s *AuthService) UserByToken(ctx context.Context, token string) (*model.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.tokens[token]
	if !ok {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

// CreateUser — создать нового пользователя.
// Вызывается ТОЛЬКО из админского endpoint'а
// Пароль хешируется SHA-256 и сохраняется в этом виде.
func (s *AuthService) CreateUser(ctx context.Context, login, password, role string) error {
	login = strings.TrimSpace(login)

	// 1. Валидация.
	if login == "" || len(login) > 50 {
		return ErrValidation
	}
	if len(password) < 6 || len(password) > 100 {
		return ErrValidation
	}
	if role != "admin" && role != "manager" {
		return ErrValidation
	}

	// 2. Логин свободен?
	exists, err := s.userRepo.ExistsByLogin(ctx, login)
	if err != nil {
		return err
	}
	if exists {
		return ErrLoginTaken
	}

	// 3. Хешируем и сохраняем.
	hash := hashPassword(password)
	_, err = s.userRepo.Create(ctx, login, hash, role)
	return err
}
