# Montis — контракт REST API

Документ для backend-разработчика. Фронтенд ходит **только** по этим эндпоинтам
(см. `src/api/index.ts`), а заглушка `src/api/mock/server.ts` повторяет это же поведение.

- Базовый префикс: `/api` (настраивается `VITE_API_URL`).
- Формат: JSON, `Content-Type: application/json`.
- Даты/время: ISO 8601 в UTC (`2026-11-18T07:00:00.000Z`). Параметр `date` у слотов — `YYYY-MM-DD` в часовом поясе курорта.
- Авторизация: `Authorization: Bearer <token>`. Токен выдают `/auth/login` и `/auth/register`.
- Деньги — целые рубли (`2500`).

## Ошибки

Любой ответ 4xx/5xx:

```json
{ "message": "Человекочитаемый текст", "errors": { "email": "Email уже зарегистрирован" } }
```

`errors` необязателен; если есть — фронт подсветит поля формы.

| Код | Когда |
|---|---|
| 400 / 422 | ошибка валидации |
| 401 | нет/просрочен токен — фронт разлогинит пользователя |
| 403 | нет прав (не та роль / чужая бронь) |
| 404 | не найдено |
| 409 | конфликт: **интервал уже занят**, email занят, не хватает снаряжения |
| 402 | (демо-оплата) платёж отклонён |

## Сущности

```ts
type Role = 'client' | 'instructor' | 'admin';
type BookingStatus = 'pending' | 'paid' | 'cancelled' | 'completed';

User        { id, name, phone, email, role }
Lesson      { id, title, description, category, price, durationMin, imageUrl?, active }
            // category: ski | snowboard | individual | group | kids | freestyle
Equipment   { id, title, description, category, pricePerHour, sizes: string[], stock, imageUrl?, active }
            // category: ski | snowboard | boots | helmet | kit | goggles
Instructor  { id, userId, name, email, specialization, categories: LessonCategory[],
              experienceYears, pricePerHour, bio, photoUrl? }
Slot        { start, end, available }
Booking     { id, type: 'lesson'|'rental', status, userId, clientName, title,
              lessonId?, instructorId?, instructorName?,      // для занятия
              equipmentId?, size?, quantity?,                 // для аренды
              start, end, total, createdAt }
Availability { id, instructorId, start, end }                  // рабочее окно инструктора
ScheduleCell { start, end, status: 'free'|'booked', availabilityId?, bookingId? }
```

Статусы: `pending` (ожидание оплаты) → `paid` → `completed`; `cancelled` — отмена до начала.

## Авторизация

| Метод | Путь | Тело | Ответ |
|---|---|---|---|
| POST | `/auth/register` | `{ name, phone, email, password }` | `201 { token, user }` (роль всегда `client`) |
| POST | `/auth/login` | `{ email, password }` | `{ token, user }` / 401 |
| POST | `/auth/logout` | — | 204 |
| GET | `/auth/me` | — | `User` |

## Публичный каталог (без токена)

| Метод | Путь | Query | Ответ |
|---|---|---|---|
| GET | `/lessons` | `category?` | `Lesson[]` (только `active`) |
| GET | `/lessons/:id` | | `Lesson` |
| GET | `/equipment` | `category?` | `Equipment[]` (только `active`) |
| GET | `/equipment/:id` | | `Equipment` |
| GET | `/equipment/:id/availability` | `start, end, size` | `{ available: number }` — сколько единиц свободно на период |
| GET | `/instructors` | `lessonId?` — только те, кто ведёт это занятие | `Instructor[]` |
| GET | `/instructors/:id` | | `Instructor` |
| GET | `/instructors/:id/slots` | `date=YYYY-MM-DD, lessonId` | `Slot[]` |
| POST | `/feedback` | `{ name, contact, message }` | 201 |

**`/slots` — главное правило ТЗ:** свободные интервалы считает бэкенд.
Слот `available: true`, если интервал `[start, start + lesson.durationMin)` целиком
покрыт рабочими окнами инструктора, не пересекается с его бронями в статусе
`pending`/`paid` и начинается в будущем. Занятые слоты тоже возвращаются (`available: false`),
фронт показывает их серыми.

## Бронирования (клиент)

| Метод | Путь | Тело | Ответ |
|---|---|---|---|
| POST | `/bookings/lesson` | `{ lessonId, instructorId, start }` | `201 Booking` (status `pending`), **409 если интервал занят** |
| POST | `/bookings/rental` | `{ equipmentId, size, quantity, start, end }` | `201 Booking`, 409 если не хватает `stock` |
| GET | `/me/bookings` | query `type?=lesson\|rental` | `Booking[]` (свои) |
| GET | `/bookings/:id` | | `Booking` (клиент — только свою) |
| POST | `/bookings/:id/cancel` | | `Booking` (только `pending/paid` и до начала) |
| POST | `/bookings/:id/pay` | `{ cardNumber, expiry, cvc }` | `Booking` (status `paid`) — **демо-оплата**, реальных платежей нет |

Сервер обязан повторно проверить свободность интервала в момент `POST /bookings/lesson`
(транзакция/блокировка), чтобы исключить двойное бронирование.
Стоимость аренды: `ceil(часы) × pricePerHour × quantity`.

## Профиль

| Метод | Путь | Тело | Ответ |
|---|---|---|---|
| PATCH | `/me/profile` | `{ name, phone, email }` | `User` |
| POST | `/me/password` | `{ currentPassword, newPassword }` | 204 / 422 `errors.currentPassword` |

## Инструктор (роль `instructor`)

| Метод | Путь | Query / тело | Ответ |
|---|---|---|---|
| GET | `/instructor/me` | | `Instructor` |
| GET | `/instructor/bookings` | `from, to` (ISO) | `Booking[]` без `cancelled` |
| GET | `/instructor/schedule` | `from, to` | `ScheduleCell[]` (часовые ячейки рабочих окон) |
| POST | `/instructor/availability` | `{ start, end }` | `201 Availability` (сервер режет на часовые окна) |
| DELETE | `/instructor/availability/:id` | | 204 / 409 если на окне есть запись |
| POST | `/bookings/:id/complete` | | `Booking` (status `completed`, только `paid`) |

## Администратор (роль `admin`)

| Метод | Путь | Тело / query | Ответ |
|---|---|---|---|
| GET | `/admin/stats` | | `{ users, instructors, bookingsToday, revenueWeek, bookingsByWeekday: number[7], latestBookings: Booking[] }` |
| GET | `/admin/users` | `role?` | `User[]` |
| PATCH | `/admin/users/:id` | `{ name?, phone?, email? }` | `User` |
| POST | `/admin/instructors` | `{ name, email, password, specialization }` | `201 Instructor` (создаёт и User c ролью `instructor`) |
| GET / POST | `/admin/lessons` | `Omit<Lesson,'id'>` | `Lesson[]` / `201 Lesson` (включая скрытые) |
| PATCH / DELETE | `/admin/lessons/:id` | `Partial<Lesson>` | `Lesson` / 204 |
| GET / POST | `/admin/equipment` | `Omit<Equipment,'id'>` | `Equipment[]` / `201 Equipment` |
| PATCH / DELETE | `/admin/equipment/:id` | `Partial<Equipment>` | `Equipment` / 204 |
| GET | `/admin/bookings` | `status?` | `Booking[]` (новые сверху) |
| PATCH | `/admin/bookings/:id` | `{ status }` | `Booking` (409 при возврате отменённой брони на занятое время) |

## Минимальная модель БД

`users` · `instructors` (1:1 users) · `lessons` · `instructor_lessons`/категории ·
`availability` (рабочие окна) · `equipment` · `bookings` (тип занятие/аренда) · `feedback`.
