package main

// Запуск из папки backend
//   go run ./cmd/cli
// сначала выполнить
//   chcp 65001

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/BlueLamp23/Sotrud/backend/internal/config"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

// reader — один на всю программу. Читает строки из stdin.
var reader = bufio.NewReader(os.Stdin)

// readLine — прочитать строку, обрезать \r\n и пробелы.
func readLine(prompt string) string {
	fmt.Print(prompt)
	s, _ := reader.ReadString('\n')
	return strings.TrimSpace(s)
}

// readInt — прочитать число. Если ввели не число — вернёт 0.
func readInt(prompt string) int64 {
	s := readLine(prompt)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// positionTitle — название текущей должности для вывода.
// Если должности нет (нет активного назначения) — "(нет)".
func positionTitle(title *string) string {
	if title == nil || *title == "" {
		return "(нет)"
	}
	return *title
}

func main() {
	// Настройки и база — как в сервере.
	cfg := config.Load()
	db, err := repository.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()

	//Репозитории и сервисы — те же, что в main.go сервера.
	positionRepo := repository.NewPositionRepo(db)
	employeeRepo := repository.NewEmployeeRepo(db)
	appointmentRepo := repository.NewAppointmentRepo(db)

	positionSvc := service.NewPositionService(positionRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo)
	appointmentSvc := service.NewAppointmentService(appointmentRepo, employeeRepo, positionRepo)

	ctx := context.Background()

	// 3. Цикл меню.
	for {
		fmt.Println()
		fmt.Println("=== Учёт сотрудников и должностей ===")
		fmt.Println("1. Список сотрудников (с должностями)")
		fmt.Println("2. Поиск сотрудника по ФИО")
		fmt.Println("3. Список должностей")
		fmt.Println("4. Добавить сотрудника")
		fmt.Println("5. Добавить должность")
		fmt.Println("6. Назначить на должность")
		fmt.Println("7. Снять с должности")
		fmt.Println("8. История назначений сотрудника")
		fmt.Println("9. Удалить должность")
		fmt.Println("10. Удалить сотрудника")
		fmt.Println("0. Выход")
		fmt.Println()

		choice := readLine("Выберите пункт: ")

		switch choice {
		case "1":
			listEmployees(ctx, employeeSvc)
		case "2":
			searchEmployees(ctx, employeeSvc)
		case "3":
			listPositions(ctx, positionSvc)
		case "4":
			addEmployee(ctx, employeeSvc)
		case "5":
			addPosition(ctx, positionSvc)
		case "6":
			assign(ctx, appointmentSvc)
		case "7":
			unassign(ctx, appointmentSvc)
		case "8":
			history(ctx, appointmentSvc)
		case "9":
			deletePosition(ctx, positionSvc)
		case "10":
			deleteEmployee(ctx, employeeSvc)
		case "0":
			fmt.Println("До встречи.")
			return
		default:
			fmt.Println("Нет такого пункта.")
		}
	}
}

func listEmployees(ctx context.Context, svc *service.EmployeeService) {
	list, err := svc.List(ctx, "", "", 0)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	if len(list) == 0 {
		fmt.Println("Сотрудников нет.")
		return
	}

	fmt.Printf("%-4s %-30s %-20s %-10s\n", "ID", "ФИО", "Должность", "Статус")
	fmt.Println(strings.Repeat("-", 70))
	for _, e := range list {
		fmt.Printf("%-4d %-30s %-20s %-10s\n", e.EmployeeID, e.FullName, positionTitle(e.PositionTitle), e.EmployeeStatus)
	}
}

func searchEmployees(ctx context.Context, svc *service.EmployeeService) {
	q := readLine("Подстрока ФИО: ")
	if q == "" {
		return
	}
	list, err := svc.List(ctx, q, "", 0)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	if len(list) == 0 {
		fmt.Println("Никого не найдено.")
		return
	}
	for _, e := range list {
		fmt.Printf("[%d] %s — %s (%s)\n", e.EmployeeID, e.FullName, positionTitle(e.PositionTitle), e.EmployeeStatus)
	}
}

func listPositions(ctx context.Context, svc *service.PositionService) {
	list, err := svc.List(ctx)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	if len(list) == 0 {
		fmt.Println("Должностей нет.")
		return
	}

	fmt.Printf("%-4s %-25s %-10s\n", "ID", "Название", "Активна")
	fmt.Println(strings.Repeat("-", 45))
	for _, p := range list {
		active := "нет"
		if p.IsActive {
			active = "да"
		}
		fmt.Printf("%-4d %-25s %-10s\n", p.ID, p.Title, active)
	}
}

func addEmployee(ctx context.Context, svc *service.EmployeeService) {
	fullName := readLine("ФИО: ")
	phone := readLine("Телефон (можно пусто): ")
	email := readLine("Email (можно пусто): ")

	status := readLine("Статус (1 - Работает, 2 - Уволен): ")
	st := service.StatusWorking
	if status == "2" {
		st = service.StatusFired
	}

	id, err := svc.Create(ctx, fullName, phone, email, st)
	if err != nil {
		// Сюда попадут ErrValidation, ErrEmailTaken — с понятным текстом.
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Создан сотрудник с id =", id)
}

func addPosition(ctx context.Context, svc *service.PositionService) {
	title := readLine("Название должности: ")
	activeStr := readLine("Активна? (1 - да, 0 - нет): ")
	active := activeStr != "0"

	id, err := svc.Create(ctx, title, active)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Создана должность с id =", id)
}

func assign(ctx context.Context, svc *service.AppointmentService) {
	empID := readInt("ID сотрудника: ")
	posID := readInt("ID должности: ")
	start := readLine("Дата начала (YYYY-MM-DD): ")

	id, err := svc.Assign(ctx, empID, posID, start)
	if err != nil {
		// ErrAlreadyAssigned — «уже есть активное назначение».
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Назначение создано, id =", id)
}

func unassign(ctx context.Context, svc *service.AppointmentService) {
	empID := readInt("ID сотрудника: ")
	end := readLine("Дата окончания (YYYY-MM-DD): ")

	if err := svc.Unassign(ctx, empID, end); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Снят с должности.")
}

func history(ctx context.Context, svc *service.AppointmentService) {
	empID := readInt("ID сотрудника: ")
	list, err := svc.History(ctx, empID)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	if len(list) == 0 {
		fmt.Println("История пуста.")
		return
	}

	fmt.Printf("%-4s %-25s %-12s %-12s\n", "ID", "Должность", "Начало", "Окончание")
	fmt.Println(strings.Repeat("-", 60))
	for _, a := range list {
		end := "сейчас"
		if a.EndDate != nil {
			end = *a.EndDate
		}
		fmt.Printf("%-4d %-25s %-12s %-12s\n", a.ID, a.PositionTitle, a.StartDate, end)
	}
}

func deletePosition(ctx context.Context, svc *service.PositionService) {
	id := readInt("ID должности: ")
	if err := svc.Delete(ctx, id); err != nil {
		// ErrPositionInUse — «на должности есть сотрудники».
		// Это то же правило, что 409 в HTTP-версии.
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Удалено.")
}

func deleteEmployee(ctx context.Context, svc *service.EmployeeService) {
	id := readInt("ID сотрудника: ")
	if err := svc.Delete(ctx, id); err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Println("Удалено.")
}
