package handler

// HTTP-обёртки для должностей.

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

type PositionHandler struct {
	svc *service.PositionService
}

func NewPositionHandler(svc *service.PositionService) *PositionHandler {
	return &PositionHandler{svc: svc}
}

// GET

// List — отдать все должности.
// Тело запроса не нужно, параметров нет.
func (h *PositionHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// POST

// positionRequest — тело для создания и обновления
type positionRequest struct {
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
}

// Create — создать должность.
func (h *PositionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req positionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	id, err := h.svc.Create(r.Context(), req.Title, req.IsActive)
	if err != nil {
		writeServiceError(w, err) // ErrValidation → 400
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT

// Update — изменить должность.
func (h *PositionHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. Разобрать id из URL.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// 2. Разобрать тело.
	var req positionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// 3. Позвать сервис.
	if err := h.svc.Update(r.Context(), id, req.Title, req.IsActive); err != nil {
		writeServiceError(w, err) // ErrValidation → 400, ErrNotFound → 404
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE

// Delete — удалить должность.
// Если на должности есть сотрудники — сервис вернёт ErrPositionInUse, потом 409
func (h *PositionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err) // ErrPositionInUse → 409, ErrNotFound → 404
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204
}
