PRAGMA foreign_keys = ON;

BEGIN;

DROP TABLE IF EXISTS appointments;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS users;

-- 1. Таблица пользователей (Администраторы)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL, -- Тут будет хэш
    role TEXT NOT NULL DEFAULT 'admin'
);

-- 2. Таблица сотрудников
CREATE TABLE employees (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    phone TEXT,
    email TEXT UNIQUE,
    status TEXT NOT NULL DEFAULT 'Работает' CHECK (status IN ('Работает', 'Уволен'))
);

-- 3. Справочник должностей
CREATE TABLE positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL UNIQUE,
    is_active INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)) -- 1 = активна, 0 = неактивна
);

-- 4. Таблица назначений (история)
CREATE TABLE appointments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    employee_id INTEGER NOT NULL,
    position_id INTEGER NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT, -- Если NULL, значит сотрудник занимает должность по сей день
    
    -- Проверка, что дата окончания не раньше даты начала
    CHECK (end_date IS NULL OR end_date >= start_date),
    
    -- Внешние ключи
    FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE,
    FOREIGN KEY (position_id) REFERENCES positions(id) ON DELETE RESTRICT
);

-- Индексы для ускорения поиска 
CREATE INDEX idx_appointments_employee ON appointments(employee_id);
CREATE INDEX idx_appointments_position ON appointments(position_id);


-- Тестовые данные
-- Добавляем администратора
INSERT INTO users (login, password_hash, role) VALUES 
('admin', 'secret_hash_123', 'admin');

-- Добавляем сотрудников (один уволен, остальные работают)
INSERT INTO employees (full_name, phone, email, status) VALUES 
('Иванов Иван Иванович', '+79001112233', 'ivanov@mail.ru', 'Работает'),
('Смирнов Сергей Сергеевич', '+79004445566', 'smirnov@mail.ru', 'Работает'),
('Петров Петр Петрович', '+79007778899', 'petrov@mail.ru', 'Уволен'),
('Сидоров Сидор Сидорович', '+79002223344', 'sidorov@mail.ru', 'Работает');

-- Добавляем должности (Стажер неактивен, остальные активны)
INSERT INTO positions (title, is_active) VALUES 
('Инженер', 1),
('Менеджер', 1),
('Директор', 1),
('Стажер', 0);

-- Добавляем назначения 
INSERT INTO appointments (employee_id, position_id, start_date, end_date) VALUES 

(1, 4, '2019-01-10', '2020-08-31'),
(1, 1, '2020-09-01', NULL),

(2, 2, '2026-08-02', NULL),

(3, 1, '2018-05-10', '2023-11-15'),

(4, 3, '2022-01-15', NULL);

COMMIT;