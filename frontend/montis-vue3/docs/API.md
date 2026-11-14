# Предварительный API-контракт Montis

Это **предложение frontend-разработчика**, не описание существующего backend. Backend-разработчик должен согласовать и реализовать маршруты/DTO. Базовый префикс `/api`.

Все защищённые запросы: `Authorization: Bearer <token>`. Ответы JSON, ошибки `{ "message": "...", "errors": { "field": "..." } }`.

## Авторизация
- `POST /auth/register` — `{name, phone, email, password}` → `{token, user}`
- `POST /auth/login` — `{email, password}` → `{token, user}`
- `GET /auth/me` → `User`
- `POST /auth/logout` → `204`

## Каталог и доступность
- `GET /lessons?category=` → `Lesson[]`; `GET /lessons/:id` → `Lesson`
- `GET /equipment?category=` → `Equipment[]`; `GET /equipment/:id` → `Equipment`
- `GET /instructors?lessonId=` → `Instructor[]`; `GET /instructors/:id` → `Instructor`
- `GET /instructors/:id/slots?date=YYYY-MM-DD&lessonId=ID` → `[{start,end}]` (ISO 8601)
- `GET /equipment/:id/availability?start=ISO&end=ISO&size=...` → `{available:number}`

**Слоты, остатки и отсутствие двойной записи вычисляет backend.** Frontend не генерирует время самостоятельно.

## Бронирования и профиль
- `POST /bookings/lesson` — `{lessonId,instructorId,start}` → `Booking`
- `POST /bookings/rental` — `{equipmentId,size,quantity,start,end}` → `Booking`
- `GET /bookings/:id` → `Booking`; `GET /me/bookings?type=` → `Booking[]`
- `POST /bookings/:id/cancel` → `Booking`
- `POST /bookings/:id/pay` → `Booking` (временный контракт, согласовать реальную оплату)
- `PATCH /me/profile` — `{name,phone,email}` → `User`
- `POST /me/password` — `{current,next}` → `204`
- `POST /feedback` — `{name,email,message}` → `204`

## Инструктор
- `GET /instructor/me` → `Instructor`
- `GET /instructor/bookings?from=&to=` → `Booking[]`
- `GET /instructor/schedule?from=&to=` → `ScheduleCell[]`
- `POST /instructor/availability` — `{start,end}` → `Availability`
- `DELETE /instructor/availability/:id` → `204`
- `POST /bookings/:id/complete` → `Booking`

## Администратор
- `GET /admin/stats` → `{users,bookings,lessons,equipment}`
- `GET /admin/users?role=` → `User[]`; `PATCH /admin/users/:id` → `User`
- `POST /admin/instructors` — `{name,email,password,specialization}` → `Instructor`
- `GET /admin/lessons`, `POST /admin/lessons`, `PATCH /admin/lessons/:id`, `DELETE /admin/lessons/:id`
- `GET /admin/equipment`, `POST /admin/equipment`, `PATCH /admin/equipment/:id`, `DELETE /admin/equipment/:id`
- `GET /admin/bookings?status=` → `Booking[]`
- `PATCH /admin/bookings/:id` — `{status}` → `Booking`

Типы данных — `src/types/index.ts`. Обсудить enum статусов (`pending`, `paid`, `cancelled`, `completed`), валюту, часовые пояса, поведение оплаты и формат пагинации.

## Ожидаемые ошибки
`400/422` — валидация; `401` — нет авторизации; `403` — нет прав; `404` — не найдено; `409` — слот уже занят/конфликт; `500` — ошибка сервера. Не возвращать `200` на конфликт бронирования.
