package handler

// Приёмщик входа (HTTP-обёртки).

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// AuthHandler — «приёмщик входа».
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login — POST /api/login (публичный).
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	token, err := h.svc.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		writeServiceError(w, err) // ErrInvalidCredentials → 401
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// Logout — POST /api/logout (защищённый).
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && parts[0] == "Bearer" {
		h.svc.Logout(parts[1])
	}
	w.WriteHeader(http.StatusNoContent)
}

type createUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// CreateUser — POST /api/register (защищённый, только admin).
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.svc.CreateUser(r.Context(), req.Login, req.Password, req.Role); err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}
