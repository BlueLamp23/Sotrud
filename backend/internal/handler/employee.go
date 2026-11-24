package handler

// HTTP-обёртки для сотрудников.
import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// EmployeeHandler — «приёмщик сотрудников».
type EmployeeHandler struct {
	svc *service.EmployeeService
}

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// GET

// List — список сотрудников с текущими должностями.
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	// r.URL.Query() парсит часть URL после "?" в map[string][]string.
	// .Get("search") — берёт первое значение или "" если нет.
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	// position_id — необязательный фильтр по текущей должности.
	// Пустая строка или мусор -> 0 (значит «не фильтровать»).
	var positionID int64
	if raw := r.URL.Query().Get("position_id"); raw != "" {
		if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
			positionID = v
		}
	}

	list, err := h.svc.List(r.Context(), search, status, positionID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// -GET

// Get — один сотрудник по id.
func (h *EmployeeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	emp, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err) // ErrNotFound → 404
		return
	}

	writeJSON(w, http.StatusOK, emp)
}

// POST
// employeeRequest — тело для создания/обновления.
// Все поля строки, потому что phone и email могут быть NULL,
// Сервис сам решит, писать ли NULL в базу.
type employeeRequest struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// Create — создать сотрудника.
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	id, err := h.svc.Create(r.Context(), req.FullName, req.Phone, req.Email, req.Status)
	if err != nil {
		writeServiceError(w, err) // ErrValidation → 400, ErrEmailTaken → 409
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// UT
// Update — изменить сотрудника.
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req employeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.svc.Update(r.Context(), id, req.FullName, req.Phone, req.Email, req.Status); err != nil {
		writeServiceError(w, err) // ErrValidation → 400, ErrEmailTaken → 409, ErrNotFound → 404
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE

// Delete — удалить сотрудника.
// В init.sql ON DELETE CASCADE: назначения удалятся вместе с ним.
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err) // ErrNotFound → 404
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
