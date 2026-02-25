# GoProject - API управления подразделениями и сотрудниками

REST API для управления базой данных. Проект реализован на Go с использованием GORM, PostgreSQL и goose для миграций.

## Содержание

- [Технологии](#технологии)
- [Запуск проекта](#запуск-проекта)
- [API Endpoints](#api-endpoints)
- [Параметры запросов](#параметры-запросов)
- [Структура базы данных](#структура-базы-данных)

## Технологии

- **Go** 1.26
- **PostgreSQL** 18
- **GORM** - ORM для работы с БД
- **goose** - миграции базы данных
- **Docker** & **Docker Compose** - контейнеризация

## Запуск проекта

```bash
# Создание билда
docker compose up --build -d

# Запуск образа
docker compose up -d
```

## API Endpoints

## Отделы
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/departments` | Создание нового подразделения |
| GET	| `/departments/{id}`	| Получение информации об подразделении |
| PATCH	| `/departments/{id}`	| Перемещение/переименование подразделения |
| DELETE | `/departments/{id}` | Удаление подразделения |

## Сотрудники
| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/departments/{id}/employees` | Создание сотрудника в подразделение|

## Параметры запросов
| Метод | Endpoint | Параметры пути | Query параметры | Body параметры |
|-------|----------|----------------|-----------------|----------------|
| POST | `/departments` | - | - | `name`, `parent_id?` |
| POST | `/departments/{id}/employees` | `id` | - | `full_name`, `position`, `hired_at?` |
| GET | `/departments/{id}` | `id` | `depth?=1`, `include_employees?=true` | - |
| PATCH | `/departments/{id}` | `id` | - | `name?`, `parent_id?` |
| DELETE | `/departments/{id}` | `id` | `mode`, `reassign_to_department_id?` | - |

*`?` - опциональный параметр*

## Структура базы данных
**departments**
| Поле | Тип | Описание |
|------|-----|----------|
| id | uint | PRIMARY KEY |
| name | string | Название (unique в ветке) |
| parent_id | uint | FOREIGN KEY (self) |
| created_at | timestamp | Дата создания |

**employees**
| Поле | Тип | Описание |
|------|-----|----------|
| id | uint | PRIMARY KEY |
| department_id | uint | FOREIGN KEY |
| full_name | string | Полное имя |
| position | string | Должность |
| hired_at | timestamp | Дата найма |
| created_at | timestamp | Дата создания |

