package main

// Точка входа сервера
// config → db → repos → services → handlers → router → ListenAndServe

import (
	"log"
	"net/http"

	"github.com/BlueLamp23/Sotrud/backend/internal/config"
	"github.com/BlueLamp23/Sotrud/backend/internal/handler"
	"github.com/BlueLamp23/Sotrud/backend/internal/repository"
	"github.com/BlueLamp23/Sotrud/backend/internal/service"
)

func main() {
	//  Настройки
	cfg := config.Load()
	log.Printf("Config: DB=%s, Port=%s", cfg.DBPath, cfg.Port)
	db, err := repository.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("cannot open db: %v", err)
	}
	defer db.Close()
	log.Println("DB connected")

	// Репозитории
	// Каждый получает db — единственную зависимость.
	userRepo := repository.NewUserRepo(db)
	positionRepo := repository.NewPositionRepo(db)
	employeeRepo := repository.NewEmployeeRepo(db)
	appointmentRepo := repository.NewAppointmentRepo(db)

	//  Сервисы
	authSvc := service.NewAuthService(userRepo)
	positionSvc := service.NewPositionService(positionRepo)
	employeeSvc := service.NewEmployeeService(employeeRepo)
	appointmentSvc := service.NewAppointmentService(appointmentRepo, employeeRepo, positionRepo)

	// HandlerЫ
	authH := handler.NewAuthHandler(authSvc)
	positionH := handler.NewPositionHandler(positionSvc)
	employeeH := handler.NewEmployeeHandler(employeeSvc)
	appointmentH := handler.NewAppointmentHandler(appointmentSvc)

	// Роутер
	router := handler.NewRouter(authH, positionH, employeeH, appointmentH, authSvc)

	// Сервер
	addr := ":" + cfg.Port
	log.Printf("Server started: http://localhost:%s", cfg.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
