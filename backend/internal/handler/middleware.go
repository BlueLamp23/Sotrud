package handler

// Middleware — «обёртки» вокруг HTTP-handler'ов.
// Каждая обёртка: http.Handler → http.Handler.
// Это позволяет цеплять их друг за друга:
//   Recoverer(Logger(CORS(router)))
// Запрос проходит через каждую по очереди.

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// CORS разрешает фронтенду (localhost:5173) обращаться к API (localhost:8080).
// Без этого браузер заблокирует запросы из другого источника
//
//	Access-Control-Allow-Origin - с какого адреса можно
//	Access-Control-Allow-Methods - какие HTTP-методы
//	Access-Control-Allow-Headers - какие заголовки
//
// OPTIONS — preflight-запрос.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // * = любой origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Logger
// responseWriter-обёртка, чтобы запомнить код ответа (Handler его не возвращает)
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger пишет в консоль: метод, путь, код, время выполнения.
// Пример: POST /api/login 200 3.2ms
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)

		log.Printf("%s %s %d %s",
			r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

// Recoverer

// ловит panic в handler'ах, с panic в одном запросе = 500 этому запросу, остальные работают.
// recover() работает только внутри defer, иначе возвращает nil.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC: %v", rec)
				http.Error(w, `{"error":"internal server error"}`,
					http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequireAuth

// contextKey — свой тип для ключей в context.
// Нельзя использовать обычную строку: два пакета могут положить по ключу "user" разные вещи и перезаписать друг друга.
type contextKey string

const userContextKey contextKey = "user"

// RequireAuth проверяет токен в заголовке Authorization.
// Если токена нет / он неверный — 401 Unauthorized.
// Если есть — кладём пользователя в context и пускаем дальше.
func RequireAuth(authSvc *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Достаём заголовок.
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"error":"missing token"}`, http.StatusUnauthorized)
				return
			}

			// Ожидаем
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid token format"}`,
					http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// Проверяем токен через сервис
			user, err := authSvc.UserByToken(r.Context(), token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`,
					http.StatusUnauthorized)
				return
			}

			// Кладём пользователя в context и идём дальше
			// handler достанет его через UserFromContext
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext - достать пользователя, которого RequireAuth положил в context.
// Возвращает nil, если middleware не сработал (например, публичный endpoint).
func UserFromContext(ctx context.Context) *model.User {
	user, _ := ctx.Value(userContextKey).(*model.User)
	return user
}

//	RequireRole пускает только пользователей с определённой ролью.
//
// Ставится ПОСЛЕ RequireAuth (иначе в context ещё нет пользователя).
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				// RequireAuth не сработал — это баг конфигурации роутера.
				http.Error(w, `{"error":"unauthorized"}`,
					http.StatusUnauthorized)
				return
			}
			if user.Role != role {
				http.Error(w, `{"error":"forbidden"}`,
					http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r) //кстати w от writer, a r от request
		})
	}
}
