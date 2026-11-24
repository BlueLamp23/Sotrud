package model

import "database/sql"

// User - учетная запись администратора
type User struct {
	ID           int64  `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

// Employee - карточка сотрудника
// Phone и Email могут быть NULL: в JSON они либо строка, либо отсутствуют.
type Employee struct {
	ID       int64   `json:"id"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone,omitempty"`
	Email    *string `json:"email,omitempty"`
	Status   string  `json:"status"`
}

// Position - справочник должностей
type Position struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"` // В скрипте замена bool 0/1
}

// Appointment - история назначений
// EndDate = NULL значит назначение ещё активно (в JSON поля не будет).
type Appointment struct {
	ID         int64   `json:"id"`
	EmployeeID int64   `json:"employee_id"`
	PositionID int64   `json:"position_id"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

// EmployeeWithPosition - структура для отображения на главной странице
// (объединяет данные из трех таблиц)
type EmployeeWithPosition struct {
	EmployeeID     int64   `json:"employee_id"`
	FullName       string  `json:"full_name"`
	EmployeeStatus string  `json:"employee_status"`
	PositionTitle  *string `json:"position_title,omitempty"`
	StartDate      *string `json:"start_date,omitempty"`
}

// AppointmentWithPosition - для страницы «История назначений».
// Включает название должности, чтобы фронт не делал повторный запрос.
type AppointmentWithPosition struct {
	ID            int64   `json:"id"`
	PositionTitle string  `json:"position_title"`
	StartDate     string  `json:"start_date"`
	EndDate       *string `json:"end_date,omitempty"`
}

// nullToStringPtr превращает значение из базы в *string:
// NULL или пустая строка -> nil (в JSON поля не будет), иначе указатель на строку.
func nullToStringPtr(s sql.NullString) *string {
	if !s.Valid || s.String == "" {
		return nil
	}
	v := s.String
	return &v
}

// StringPtrToNull - обратное преобразование: *string -> sql.NullString.
// nil -> NULL в базе.
func StringPtrToNull(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// ScanEmployee сканирует строку (id, full_name, phone, email, status) в Employee.
func (e *Employee) ScanEmployee(id int64, fullName string, phone, email sql.NullString, status string) {
	e.ID = id
	e.FullName = fullName
	e.Phone = nullToStringPtr(phone)
	e.Email = nullToStringPtr(email)
	e.Status = status
}

// ScanEmployeeWithPosition сканирует строку (employee_id, full_name, status, title, start_date).
func (e *EmployeeWithPosition) ScanEmployeeWithPosition(employeeID int64, fullName, status string, title, startDate sql.NullString) {
	e.EmployeeID = employeeID
	e.FullName = fullName
	e.EmployeeStatus = status
	e.PositionTitle = nullToStringPtr(title)
	e.StartDate = nullToStringPtr(startDate)
}

// ScanAppointment сканирует строку (id, employee_id, position_id, start_date, end_date).
func (a *Appointment) ScanAppointment(id, employeeID, positionID int64, startDate string, endDate sql.NullString) {
	a.ID = id
	a.EmployeeID = employeeID
	a.PositionID = positionID
	a.StartDate = startDate
	a.EndDate = nullToStringPtr(endDate)
}

// ScanAppointmentWithPosition сканирует строку (id, position_title, start_date, end_date).
func (a *AppointmentWithPosition) ScanAppointmentWithPosition(id int64, positionTitle, startDate string, endDate sql.NullString) {
	a.ID = id
	a.PositionTitle = positionTitle
	a.StartDate = startDate
	a.EndDate = nullToStringPtr(endDate)
}
