package handler

// Router единственное место, где перечислены все адреса.

import (
	"net/http"

	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// chain — применить несколько middleware к одному handler.
// Порядок: последний в списке срабатывает ПЕРВЫМ.
// Пример: chain(h, RequireAuth, RequireRole("admin")) → запрос: RequireAuth → RequireRole → h
func chain(h http.HandlerFunc, mws ...func(http.Handler) http.Handler) http.Handler {
	var result http.Handler = h
	// Идём с конца, чтобы первый в списке оказался внешним.
	for i := len(mws) - 1; i >= 0; i-- {
		result = mws[i](result)
	}
	return result
}

// NewRouter собирает все адреса.
// Принимает все handler'ы + AuthService (нужен для middleware RequireAuth).
func NewRouter(
	authH *AuthHandler,
	posH *PositionHandler,
	empH *EmployeeHandler,
	appH *AppointmentHandler,
	authSvc *service.AuthService,
) http.Handler {
	mux := http.NewServeMux()

	// Публичные адреса
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Вход — единственный публичный рабочий endpoint.
	mux.HandleFunc("POST /api/login", authH.Login)

	//  Защищённые адреса

	requireAuth := RequireAuth(authSvc)
	requireAdmin := RequireRole("admin")

	// Выход — нужен токен (чтобы было что удалять), но роль не важна.
	mux.Handle("POST /api/logout", chain(authH.Logout, requireAuth))

	// Создание пользователя — только админ (регистрация не публичная).
	mux.Handle("POST /api/register",
		chain(authH.CreateUser, requireAuth, requireAdmin))

	// ---- Должности ----

	// Просмотр — любой авторизованный (и админ, и менеджер).
	mux.Handle("GET /api/positions",
		chain(posH.List, requireAuth))

	// Создание/изменение/удаление — только админ.
	mux.Handle("POST /api/positions",
		chain(posH.Create, requireAuth, requireAdmin))
	mux.Handle("PUT /api/positions/{id}",
		chain(posH.Update, requireAuth, requireAdmin))
	mux.Handle("DELETE /api/positions/{id}",
		chain(posH.Delete, requireAuth, requireAdmin))

	// ---- Сотрудники ----

	// Просмотр — любой авторизованный.
	mux.Handle("GET /api/employees",
		chain(empH.List, requireAuth))
	mux.Handle("GET /api/employees/{id}",
		chain(empH.Get, requireAuth))

	// Изменения — только админ.
	mux.Handle("POST /api/employees",
		chain(empH.Create, requireAuth, requireAdmin))
	mux.Handle("PUT /api/employees/{id}",
		chain(empH.Update, requireAuth, requireAdmin))
	mux.Handle("DELETE /api/employees/{id}",
		chain(empH.Delete, requireAuth, requireAdmin))

	// ---- Назначения ----

	// История — любой авторизованный (менеджер тоже смотрит).
	mux.Handle("GET /api/employees/{employee_id}/history",
		chain(appH.History, requireAuth))

	// Назначить / снять — только админ.
	mux.Handle("POST /api/appointments",
		chain(appH.Assign, requireAuth, requireAdmin))
	mux.Handle("POST /api/appointments/{employee_id}/unassign",
		chain(appH.Unassign, requireAuth, requireAdmin))

	// ---- Статика: всё, что не /api/... ----
	// Раздаём собранный фронт из frontend/dist.
	// Пока фронта нет — будет 404, это нормально.
	mux.Handle("GET /", http.FileServer(http.Dir("../frontend/dist")))

	// ---- Обёртка middleware на ВСЁ ----
	// CORS — первым: он должен отвечать на OPTIONS раньше, чем mux
	// увидит запрос (у mux нет OPTIONS-маршрутов, вернёт 405).
	return Recoverer(Logger(CORS(mux)))
}
