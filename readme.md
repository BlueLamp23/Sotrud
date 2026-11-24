# Система учёта сотрудников и должностей

Веб-приложение для автоматизации учёта сотрудников и должностей небольшой организации. Система позволяет директору (администратору) вести учёт через удобный веб-интерфейс: добавлять сотрудников, поддерживать справочник должностей, назначать сотрудников на должности и отслеживать историю их перемещений.

## 🎯 Цели и возможности
Система создана для избавления от ошибок ручного учёта и централизации данных.

**Основной функционал:**
* **Авторизация:** Вход в систему для администратора (гость может только открыть страницу входа).
* **Управление сотрудниками:** Добавление, изменение, удаление и поиск сотрудников (ФИО, телефон, email, статус).
* **Управление должностями:** Ведение справочника должностей (название, активность).
* **Кадровый учёт:** Назначение сотрудника на должность и снятие с должности с фиксацией дат.
* **История назначений:** Просмотр истории перемещений конкретного сотрудника.
* **Экспорт данных:** Выгрузка актуальных данных и отчётности в формат JSON.
* **Валидация:** Автоматическая проверка данных на сервере (пустые поля, дубликаты, активность должности).

## 🛠 Технологический стек
* **Backend (Сервер / Бизнес-логика):** Go
* **Frontend (Клиент / Интерфейс):** Vue.js (Vite)
* **База данных:** SQLite
* **Архитектура:** Клиент-серверная, трёхуровневая (Presentation, Business Logic, Data layers).

## 📂 Структура проекта
Проект организован с  разделением на клиентскую и серверную части, а также выделением слоёв архитектуры.

> Git не отслеживает пустые папки. Если вы создаёте папки `handler/`, `service/` и т.д. и они пока пустые, добавьте в них файл `.gitkeep`, чтобы структура сохранилась в репозитории. Lock-файлы (`go.sum`, `package-lock.json`) и `.env.example` **нужно коммитить**.

```text
Sotrud/
│
├── .gitignore                                
├── .gitkeep                                   
├── .editorconfig                              
├── .env.example                              
├── Makefile                                  
├── docker-compose.yml                        
├── LICENSE                                  
├── readme.md                                 
│                                               
├── sqlite3.exe                               
├── ИНСТРУКЦИЯ.md                             
│
├── backend/
│   ├── go.mod                                
│   ├── go.sum                                 
│   ├── .gitkeep                               
│   ├── cmd/
│   │   ├── .gitkeep                          
│   │   ├── cl/                                
│   │   ├── cli/
│   │   │   └── main.go                        
│   │   └── server/
│   │       ├── .gitkeep                      
│   │       └── main.go                      
│   ├── migrations/                            
│   └── internal/
│       ├── .gitkeep                          
│       ├── config/
│       │   ├── .gitkeep                      
│       │   └── config.go                     
│       ├── model/
│       │   ├── .gitkeep                     
│       │   └── models.go                    
│       ├── repository/
│       │   ├── .gitkeep                       
│       │   ├── db.go                          
│       │   ├── errors.go                      
│       │   ├── user_repo.go                  
│       │   ├── employee_repo.go              
│       │   ├── position_repo.go               
│       │   └── appointment_repo.go            
│       ├── service/
│       │   ├── .gitkeep                   
│       │   ├── errors.go                      
│       │   ├── auth.go                        
│       │   ├── employee.go                    
│       │   ├── position.go                    
│       │   ├── appointment.go                 
│       │   └── export.go                      
│       └── handler/
│           ├── .gitkeep                   
│           ├── helpers.go                    
│           ├── middleware.go                  
│           ├── router.go                      
│           ├── auth.go                       
│           ├── employee.go                 
│           ├── position.go                
│           ├── appointment.go
│           └── export.go                     
│
├── database/
│   ├── .gitkeep                           
│   ├── init.sql                            
│   ├── lab1.db                               
│   ├── lab1.db-wal / lab1.db-shm             
│   ├── ER_diagramma.JPG                      
│   ├── Команды.txt                            
│   └── sqlite3.exe                          
│
├── docs/
│   ├── .gitkeep                              
│   ├── Схема1.png                           
│   ├── Схема2.png                          
│   ├── Схема3.png                            
│   ├── Схема4.png                            
│   ├── Схема5.png                             
│   ├── diagrams/
│   │   └── usecase.puml                       
│   └── (Схема6.png)                        
│
├── frontend/                                 
│   ├── .gitkeep                             
│   ├── package.json                      
│   ├── package-lock.json                     
│   ├── index.html                            
│   ├── vite.config.js                         
│   ├── public/.gitkeep                       
│   └── src/
│       ├── App.vue                          
│       ├── main.js                          
│       ├── api/{/.gitkeep, client.js, sotrud.js}      
│       ├── assets/{/.gitkeep, style.css}           
│       ├── components/{/.gitkeep, AppHeader, EmployeeForm, PositionForm,
│       │                AppointmentForm, DataTable, SearchBar,
│       │                ConfirmModal, ErrorBanner}.vue  
│       ├── router/{/.gitkeep, index.js}                
│       ├── stores/{/.gitkeep, auth.js}                  
│       └── views/{/.gitkeep, Login, Employees, EmployeeCard,
│                    Positions, Export}.vue               
│
└── out/
    └── .gitkeep                              
