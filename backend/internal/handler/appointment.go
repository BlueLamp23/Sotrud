package handler

// HTTP-обёртки для назначений
import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BlueLamp23/Sotrud/backend/internal/model"
	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

type AppointmentHandler struct {
	svc *service.AppointmentService
}

func NewAppointmentHandler(svc *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{svc: svc}
}

// POST
// assignRequest — тело для назначения.
type assignRequest struct {
	EmployeeID int64  `json:"employee_id"`
	PositionID int64  `json:"position_id"`
	StartDate  string `json:"start_date"` // "YYYY-MM-DD"
}

// Assign — назначить сотрудника на должность.
// Возможные ответы:
//
//	201 — назначен, {"id": <id назначения>}
//	400 — неверная дата / невалидный JSON
//	404 — сотрудник или должность не существует
//	409 — у сотрудника уже есть активное назначение
func (h *AppointmentHandler) Assign(w http.ResponseWriter, r *http.Request) {
	var req assignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	id, err := h.svc.Assign(r.Context(), req.EmployeeID, req.PositionID, req.StartDate)
	if err != nil {
		writeServiceError(w, err) // ErrValidation → 400, ErrNotFound → 404, ErrAlreadyAssigned → 409
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// POST
// unassignRequest — тело для снятия.
type unassignRequest struct {
	EndDate string `json:"end_date"` // "YYYY-MM-DD"
}

// Unassign — снять сотрудника с должности.
// Возможные ответы:
//
//	204 — снят, тела нет
//	400 — неверная дата / end_date раньше start_date / невалидный JSON
//	404 — у сотрудника нет активного назначения (или его нет в базе)
func (h *AppointmentHandler) Unassign(w http.ResponseWriter, r *http.Request) {
	// employee_id берём из URL: /api/appointments/5/unassign
	employeeID, err := strconv.ParseInt(r.PathValue("employee_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	var req unassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.svc.Unassign(r.Context(), employeeID, req.EndDate); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET

// History — история назначений сотрудника.
// Если сотрудник не существует — вернёт пустой массив, а не ошибку
func (h *AppointmentHandler) History(w http.ResponseWriter, r *http.Request) {
	employeeID, err := strconv.ParseInt(r.PathValue("employee_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid employee_id")
		return
	}

	history, err := h.svc.History(r.Context(), employeeID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Если history == nil (го вернёт nil при пустом result)
	//с пустым проще
	if history == nil {
		history = []model.AppointmentWithPosition{}
	}

	writeJSON(w, http.StatusOK, history)
}
