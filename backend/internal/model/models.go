package model

import "database/sql"

// User — учетная запись администратора
type User struct {
	ID           int64          `json:"id"`
	Login        string         `json:"login"`
	PasswordHash string         `json:"-"`
	Role         string         `json:"role"`
	EmployeeID   sql.NullInt64  `json:"employee_id"`
}

// Employee — карточка сотрудника
type Employee struct {
	ID       int64          `json:"id"`
	FullName string         `json:"full_name"`
	Phone    sql.NullString `json:"phone"`
	Email    sql.NullString `json:"email"`
	Status   string         `json:"status"`
}

// Position — справочник должностей
type Position struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"` // В скрипте замена bool 0/1
}

// Appointment — история назначений
type Appointment struct {
	ID         int64          `json:"id"`
	EmployeeID int64          `json:"employee_id"`
	PositionID int64          `json:"position_id"`
	StartDate  string         `json:"start_date"`
	EndDate    sql.NullString `json:"end_date"`
}

// EmployeeWithPosition — структура для отображения на главной странице
// (объединяет данные из трех таблиц)
type EmployeeWithPosition struct {
	EmployeeID   int64          `json:"employee_id"`
	FullName     string         `json:"full_name"`
	EmployeeStatus string       `json:"employee_status"`
	PositionTitle sql.NullString `json:"position_title"`
	StartDate     sql.NullString `json:"start_date"`
}