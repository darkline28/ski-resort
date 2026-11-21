# MONTIS — API-контракт для Frontend

**Дополнение по последнему архиву:** JWT admin, GET `/admin/clients` возвращает массив `{id,name,email,phone}` пользователей с ролью client. Ответ 200, ошибки 401/403/500. Этот маршрут зарегистрирован в актуальном backend и используется для назначения инструктора. С учётом него в архиве 59 операций.

**Версия:** 21.11.2026 · актуальное состояние локального backend MONTIS

**Состояние:** 58 реализованных операций. Все 16 базовых и 35 ранее запланированных операций присутствуют в текущем OpenAPI и зарегистрированы в backend. Дополнительно реализованы 7 операций административного чтения и управления назначениями услуг, описанных в разделе 2.6.

**Основание:** текущие `backend/api/openapi.yaml`, сгенерированные маршруты, обработчики, сервисы и репозитории локальной ветки `feature/equipment-bookings`, включая незакоммиченные изменения. «Реализовано» означает наличие кода, а не подтверждение сквозного тестирования всех операций на реальной БД. Это контракт текущей версии, без обещаний ещё не реализованных функций.

**Адрес API:** backend слушает порт `8080`; пути ниже указаны без префикса `/api`. В режиме разработки frontend может обращаться к `/api` через Vite proxy, который удаляет этот префикс. `/api` не является маршрутом самого Go-сервера. Прямое обращение браузера с другого origin требует отдельно настроенного прокси/CORS; CORS middleware в текущем backend нет.

**Общие правила:** основной формат — JSON; время — RFC 3339/ISO 8601 с часовым поясом (например `2026-11-22T10:00:00+03:00`), даты — `YYYY-MM-DD`; суммы — целые числа в рублях. `*` у поля означает обязательное поле исходной схемы; отсутствие `*` — необязательное поле. `null` возможен только там, где он прямо указан или дополнительно оговорён у DTO. Обязательность полей приведена по текущим схемам конкретных операций.

**JWT:** для защищённых ручек заголовок `Authorization: Bearer <access_token>`. Все публичные GET в разделе 1 не требуют JWT, если отдельно не указано иное. Для создаваемых операций доступ указывается явно. `user_id` в запросах клиента не передаётся, сервер определяет его из токена.

**Ошибки бизнес-обработчиков:** обычно `{ "error": "string", "message": "string" }`. Middleware авторизации и стандартные обработчики разбора параметров/JSON могут вернуть обычный текст, а не JSON. В ошибке получения списка комплектов возможен пустой JSON-объект. Клиент должен проверять HTTP-статус и корректно обрабатывать отсутствие JSON/`message`. Точные объявленные HTTP-коды всех 58 операций приведены в разделе 5; особенности фактических ошибок — в разделе 4.

---

# 1. Базовый API — 16 операций (реализовано)

Имена полей и обязательность соответствуют текущему OpenAPI. Это базовые операции, ранее переданные frontend; дополнительные уже реализованные операции перечислены в разделе 2.

## POST `/auth/register` — Регистрация пользователя

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** {`name`*: string; `email`*: string; `phone`*: string; `password`*: string}.
- **Ответ 201:** {`access_token`*: string; `token_type`*: string(Bearer); `user`*: {`id`*: integer; `name`*: string; `email`*: string; `phone`*: string; `role`*: string(client,instructor,admin)}}.
- **Ошибки по OpenAPI:** 400, 409, 500.

## POST `/auth/login` — Вход пользователя

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** {`email`*: string; `password`*: string}.
- **Ответ 200:** {`access_token`*: string; `token_type`*: string(Bearer); `user`*: {`id`*: integer; `name`*: string; `email`*: string; `phone`: string; `role`*: string(client,instructor,admin)}}.
- **Ошибки по OpenAPI:** 400, 401, 500.

## GET `/auth/me` — Получение текущего пользователя

- **Доступ:** JWT Bearer; client / instructor / admin.
- **Параметры:** нет.
- **JSON запроса:** не требуется.
- **Ответ 200:** {`id`*: integer; `name`*: string; `email`*: string; `phone`*: string; `role`*: string(client,instructor,admin)}.
- **Ошибки по OpenAPI:** 401, 500.

## GET `/equipment` — Получить список снаряжения

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `category` (query, необязательно, string).
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `category`*: string; `price_per_hour`*: integer; `quantity`*: integer; `image_url`: string; `description`: string}].
- **Ошибки по OpenAPI:** 500.

## GET `/equipment/{id}` — Получить снаряжение по id

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer).
- **JSON запроса:** не требуется.
- **Ответ 200:** {`id`*: integer; `name`*: string; `category`*: string; `price_per_hour`*: integer; `quantity`*: integer; `image_url`: string; `description`: string}.
- **Ошибки по OpenAPI:** 400, 404, 500.

## GET `/equipment/categories` — Получить категории снаряжения

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `code`*: string}].
- **Ошибки по OpenAPI:** 500.

## GET `/equipment/sets` — Получить список комплектов снаряжения

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `description`: string; `sport_type`*: string(ski,snowboard); `age_group`*: string(adult,child); `price_per_hour`*: integer; `is_active`*: boolean}].
- **Ошибки по OpenAPI:** 500.

## GET `/equipment/sets/{id}` — Получить комплект снаряжения по id

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer).
- **JSON запроса:** не требуется.
- **Ответ 200:** {`id`*: integer; `name`*: string; `description`: string; `sport_type`*: string(ski,snowboard); `age_group`*: string(adult,child); `price_per_hour`*: integer; `is_active`*: boolean; `items`*: массив {`category_id`*: integer; `category_name`*: string; `quantity`*: integer}}.
- **Ошибки по OpenAPI:** 400, 404, 500.

## GET `/instructors` — Получить список всех инструкторов

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `specialization`*: string; `description`: string | null; `experience_years`*: integer; `is_active`*: boolean; `image_url`: string | null}].
- **Ошибки по OpenAPI:** 500.

## GET `/instructors/{id}` — Получить информацию об инструкторе

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer).
- **JSON запроса:** не требуется.
- **Ответ 200:** {`id`*: integer; `name`*: string; `specialization`*: string; `description`: string | null; `experience_years`*: integer; `is_active`*: boolean; `image_url`: string | null}.
- **Ошибки по OpenAPI:** 404, 400, 500.

## GET `/instructors/{id}/schedule` — Получить расписание инструктора по id

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer), `date` (query, обязательно, string/date).
- **JSON запроса:** не требуется.
- **Ответ 200:** {`instructor_id`*: integer; `date`*: string; `schedule`*: массив {`id`*: integer; `starts_at`*: string; `ends_at`*: string}; `booked_intervals`*: массив {`starts_at`*: string; `ends_at`*: string}}.
- **Ошибки по OpenAPI:** 404, 400, 500.

## GET `/instructors/{id}/available-slots` — Получить доступные слоты инструктора для выбранной услуги

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer), `service_id` (query, обязательно, integer), `date` (query, обязательно, string/date).
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`starts_at`*: string; `ends_at`*: string}].
- **Ошибки по OpenAPI:** 400, 404, 422, 500.

## GET `/instructors/me/bookings` — Получить бронирования занятий текущего инструктора

- **Доступ:** JWT Bearer; только instructor.
- **Параметры:** `date` (query, необязательно, string/date).
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `client_name`*: string; `service_name`*: string; `starts_at`*: string; `ends_at`*: string; `status`*: string(confirmed,cancelled,completed,no_show)}].
- **Ошибки по OpenAPI:** 401, 400, 403, 500.

## GET `/instructors/{id}/services` — Получить услуги конкретного инструктора

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer).
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `duration_minutes`*: integer; `price`*: integer; `is_active`*: boolean; `description`: string | null}].
- **Ошибки по OpenAPI:** 400, 404, 500.

## GET `/services` — Получить список доступных услуг

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** нет.
- **JSON запроса:** не требуется.
- **Ответ 200:** массив [{`id`*: integer; `name`*: string; `duration_minutes`*: integer; `price`*: integer; `is_active`*: boolean; `description`: string | null}].
- **Ошибки по OpenAPI:** 500.

## GET `/services/{id}` — Получить услугу по id

- **Доступ:** Публичный (JWT не требуется).
- **Параметры:** `id` (path, обязательно, integer).
- **JSON запроса:** не требуется.
- **Ответ 200:** {`id`*: integer; `name`*: string; `duration_minutes`*: integer; `price`*: integer; `is_active`*: boolean; `description`: string | null}.
- **Ошибки по OpenAPI:** 400, 404, 500.

### Уточнения по существующим DTO

- `Equipment`: `id`, `name`, `category`, `price_per_hour`, `quantity` обязательны; `description` и `image_url` необязательны. **Поля `sizes` и `active` в текущем ответе нет.** Значение `quantity` отражает количество по текущей реализации каталога, не гарантирует доступность в выбранный период.
- `EquipmentCategory`: обязательные `id`, `name`, `code`; примеры кодов: `ski`, `snowboard`, `jacket`, `pants`, `boots`, `helmet`, `goggles`. Фактический набор брать из API; поле `code` присутствует в БД и backend.
- `EquipmentSet`: `id`, `name`, `sport_type`, `age_group`, `price_per_hour`, `is_active` обязательны, `description` необязателен. `EquipmentSetDetail` дополнительно содержит обязательный `items`, элемент `{category_id, category_name, quantity}`.
- `InstructorScheduleResponse` — объект `{ instructor_id, date, schedule: [{id, starts_at, ends_at}], booked_intervals: [{starts_at, ends_at}] }`.
- `AvailableSlot` — `{starts_at, ends_at}`, а не `{start, end}`; начало слотов с шагом 30 минут; дата слотов определяется по московскому времени.
- `InstructorBooking` — `{id, client_name, service_name, starts_at, ends_at, status}`.
- `Service` — `id`, `name`, `duration_minutes`, `price`, `is_active` обязательны; `description` необязателен и допускает `null`. В текущем API каталог занятий — это **`/services`**, а не `/lessons`.
- Ответ авторизации `POST /auth/register` обязательно содержит `user.phone`; `POST /auth/login` содержит `user.phone` как необязательное поле. `token_type` равен `Bearer`.

---

# 2. Бронирования, аренда и администрирование — реализовано

**Все операции этого раздела реализованы в текущем backend.** Бывшие отметки «запланировано» больше не применяются. DTO, тела запросов, ответы и действующие ограничения приведены ниже.

## 2.1. Бронирования занятий

Объект `LessonBooking` (базовый ответ):

```json
{
  "id": 15, "user_id": 3, "instructor_id": 1, "service_id": 1,
  "starts_at": "2026-11-22T10:00:00+03:00",
  "ends_at": "2026-11-22T11:00:00+03:00",
  "status": "confirmed", "total_price": 2000,
  "created_at": "2026-11-20T18:00:00+03:00"
}
```

Объект `LessonBookingDetail`: все поля `LessonBooking` + `instructor_name: string`, `service_name: string`. Эти два поля обязательны в ответах списка и деталей. `total_price`, `ends_at`, `user_id`, `status` определяет backend. Для интерфейса `starts_at` следует выбирать из `/instructors/{id}/available-slots`. Backend проверяет будущий момент, расписание и отсутствие пересечения, но отдельно не проверяет принадлежность времени списку слотов с шагом 30 минут.

| Метод и маршрут | JWT / роль | Запрос / параметры | Успешный ответ | Ошибки |
|---|---|---|---|---|
| `POST /lesson-bookings` | Да / client | JSON: **обязательные** `instructor_id: integer`, `service_id: integer`, `starts_at: date-time` | `201` — `LessonBooking` | 400, 401, 403, 404, 409, 422, 500 |
| `GET /lesson-bookings/my` | Да / client | Query `status?`: `confirmed/cancelled/completed/no_show`; без фильтра все записи | `200` — массив `LessonBookingDetail[]`, `[]` если пусто | 400, 401, 403, 500 |
| `GET /lesson-bookings/{id}` | Да / client | Path `id: integer` | `200` — `LessonBookingDetail` только для владельца | 400, 401, 403, 404, 500 |
| `PATCH /lesson-bookings/{id}/cancel` | Да / client | Path `id: integer`; **тела нет** | `200` — `LessonBookingDetail`, `status: cancelled` | 400, 401, 403, 404, 409, 500 |

Правила: запись сразу `confirmed`, только на будущий свободный слот, инструктор активен и оказывает услугу. Клиент отменяет своё занятие со статусом `confirmed` только до его начала. Цена и длительность рассчитываются сервером. Перенос не включён.

## 2.2. Аренда оборудования и комплектов

`GET /equipment/availability` — **публичный**. Query: **обязательные** `starts_at`, `ends_at` (date-time); **необязательный** `equipment_id` (integer). Ответ `200` — массив `{equipment_id: integer, size: string, available_quantity: integer}[]`. Показывает доступность моделей по размеру в заданный период. Ошибки 400/404/500. Если доступного размера нет, значение `available_quantity: 0` или элемент отсутствует; фронт трактует оба случая как недоступность.

`POST /equipment-bookings` — JWT, роль `client`. В запросе обязательны `starts_at`, `ends_at` и хотя бы одна позиция в `items` или `sets`; отсутствующий массив считается пустым. Время — date-time, период в будущем, кратен целым часам (минимум один час).

```json
{
  "starts_at": "2026-11-22T10:00:00+03:00",
  "ends_at": "2026-11-22T13:00:00+03:00",
  "items": [
    {"equipment_id": 4, "size": "M", "quantity": 1}
  ],
  "sets": [
    {
      "equipment_set_id": 1,
      "quantity": 1,
      "sizes": [
        {"category_id": 1, "size": "170"},
        {"category_id": 5, "size": "42"},
        {"category_id": 6, "size": "M"}
      ]
    }
  ]
}
```

- `items[]`: обязательные `equipment_id: integer`, `size: string`, `quantity: integer >= 1`.
- `sets[]`: обязательные `equipment_set_id: integer`, `quantity: integer >= 1`, `sizes: array`; `sizes[]`: обязательные `category_id: integer`, `size: string`. Размер указывается для каждой категории комплекта. Для первой версии один элемент `sets[]` описывает комплект одной комбинации размеров; несколько размеров оформляются отдельными элементами.
- `items` и `sets` опциональны по отдельности. Оба могут присутствовать в одном заказе. Лыжные палки входят в лыжи по умолчанию и отдельно в комплекте не передаются.

`EquipmentRental` — ответ `201` от `POST` и `200` от `GET`/`PATCH`:

```json
{
  "id": 25,
  "user_id": 3,
  "starts_at": "2026-11-22T10:00:00+03:00",
  "ends_at": "2026-11-22T13:00:00+03:00",
  "status": "confirmed",
  "total_price": 6300,
  "created_at": "2026-11-20T18:00:00+03:00",
  "items": [
    {"equipment_unit_id": 12, "equipment_id": 4, "name": "Куртка", "size": "M", "booked_set_id": null, "item_total_price": 1200},
    {"equipment_unit_id": 13, "equipment_id": 1, "name": "Горные лыжи", "size": "170", "booked_set_id": 8, "item_total_price": null},
    {"equipment_unit_id": 14, "equipment_id": 5, "name": "Ботинки", "size": "42", "booked_set_id": 8, "item_total_price": null},
    {"equipment_unit_id": 15, "equipment_id": 6, "name": "Шлем", "size": "M", "booked_set_id": 8, "item_total_price": null}
  ],
  "sets": [
    {"id": 8, "equipment_set_id": 1, "name": "Лыжный комплект", "total_price": 5100}
  ]
}
```

Все поля верхнего уровня обязательны, `items` и `sets` — массивы (допускается `[]`). Каждый элемент `items` содержит обязательные `equipment_unit_id`, `equipment_id`, `name`, `size`, `booked_set_id` (целое или `null`), `item_total_price` (целое или `null`). Каждый элемент `sets`: `id`, `equipment_set_id`, `name`, `total_price`. Для элементов комплекта `item_total_price: null`: их стоимость учтена в `sets.total_price`. `id` в `sets` — ID строки забронированного комплекта, не ID его шаблона.

| Метод и маршрут | JWT / роль | Параметры | Успешный ответ | Ошибки |
|---|---|---|---|---|
| `POST /equipment-bookings` | Да / client | JSON выше | `201` — `EquipmentRental` | 400, 401, 403, 404, 409, 422, 500 |
| `GET /equipment-bookings/my` | Да / client | Query `status?`: `pending/confirmed/issued/returned/cancelled` | `200` — `EquipmentRental[]` (или `[]`) | 400, 401, 403, 500 |
| `GET /equipment-bookings/{id}` | Да / client | Path `id` | `200` — `EquipmentRental` владельца | 400, 401, 403, 404, 500 |
| `PATCH /equipment-bookings/{id}/cancel` | Да / client | Path `id`, **тела нет** | `200` — `EquipmentRental` со статусом `cancelled` | 400, 401, 403, 404, 409, 500 |

Отмена: только до начала периода аренды и при статусе `pending`/`confirmed`; после `issued` невозможна. При создании статус сразу `confirmed`.

### Конфликты

Если слот инструктора уже занят, `409 Conflict`:

```json
{"error":"booking_conflict","message":"Выбранное время недоступно"}
```

Если не хватает свободного оборудования на период, `409 Conflict`:

```json
{"error":"equipment_unavailable","message":"Выбранное оборудование недоступно на указанный период"}
```

Недопустимая отмена аренды или изменение статуса инструктором, `409 Conflict`:

```json
{"error":"invalid_status_transition","message":"Изменение статуса недоступно"}
```

Для недопустимой отмены занятия клиентом backend возвращает другой `error`:

```json
{"error":"booking_cannot_cancel","message":"Это бронирование больше нельзя отменить"}
```

Административный конфликт, включая запрещённый переход статуса и конфликт расписания:

```json
{"error":"conflict","message":"Операция конфликтует с текущим состоянием данных"}
```

При 409 показать `message` и обновить доступность. Сервер перепроверяет доступность во время создания, даже если раньше слот был свободен.

## 2.3. Статусы, кабинет инструктора

Статусы занятий: `confirmed → cancelled | completed | no_show`. Финальные: `cancelled`, `completed`, `no_show`.

Статусы аренды: `pending → confirmed | cancelled`, `confirmed → issued | cancelled`, `issued → returned`. Финальные: `cancelled`, `returned`. `pending` предусмотрен, но новые клиентские заявки сразу `confirmed`.

`PATCH /instructors/me/bookings/{id}/status` — JWT, роль `instructor`; path `id`, JSON: обязательное `status: "completed" | "no_show"`. Ответ `200` — `{id, client_name, service_name, starts_at, ends_at, status}` (тот же `InstructorBooking`, что в действующем GET). Ошибки 400/401/403/404/409/500. Инструктор может изменить только своё занятие со статусом `confirmed` после его начала. `POST /bookings/{id}/complete` **не используется**.

## 2.4. Админка — расписание и бронирования

**Все маршруты здесь и ниже:** JWT и роль `admin`; без JWT — 401, другая роль — 403. Path ID — положительное целое. Списки возвращают массивы, при отсутствии записей `[]`. Во всех ответах обновления — 200, создания — 201, удаления — 204 без тела, если в таблице не указано иное.

| Метод / маршрут | Запрос (обязательные отмечены `*`) | Ответ |
|---|---|---|
| `GET /admin/instructors/{id}/schedules` | Path `id*`; query `date_from?`, `date_to?` (YYYY-MM-DD) | `200`: `{id, starts_at, ends_at}[]` |
| `POST /admin/instructors/{id}/schedules` | Path `id*`; JSON `starts_at*`, `ends_at*` (date-time) | `201`: `{id, starts_at, ends_at}` |
| `PATCH /admin/instructors/{id}/schedules/{scheduleId}` | Path `id*`, `scheduleId*`; JSON `starts_at?`, `ends_at?`, хотя бы одно | `200`: `{id, starts_at, ends_at}` |
| `DELETE /admin/instructors/{id}/schedules/{scheduleId}` | Path `id*`, `scheduleId*`; без тела | `204` |
| `GET /admin/lesson-bookings` | Query `status?`, `date?`, `user_id?`, `instructor_id?` | `200`: `AdminLessonBooking[]` |
| `GET /admin/lesson-bookings/{id}` | Path `id*` | `200`: `AdminLessonBooking` |
| `PATCH /admin/lesson-bookings/{id}` | Path `id*`; JSON `status*`: `cancelled`, `completed` или `no_show` | `200`: `AdminLessonBooking` |
| `GET /admin/equipment-bookings` | Query `status?`, `date?`, `user_id?` | `200`: `AdminEquipmentRental[]` |
| `GET /admin/equipment-bookings/{id}` | Path `id*` | `200`: `AdminEquipmentRental` |
| `PATCH /admin/equipment-bookings/{id}` | Path `id*`; JSON `status*`: `confirmed`, `cancelled`, `issued` или `returned` | `200`: `AdminEquipmentRental` |

`AdminLessonBooking` = `LessonBookingDetail` с обязательным `client_name: string`, то есть `{id,user_id,client_name,instructor_id,instructor_name,service_id,service_name,starts_at,ends_at,status,total_price,created_at}`.

`AdminEquipmentRental` = `EquipmentRental` с обязательным `client_name: string`; вложенные `items` и `sets` те же.

Администратор может отменять подтверждённые занятия до начала, отмечать `completed`/`no_show` после начала. Для аренды разрешены переходы по общей схеме выше. `PATCH /admin/bookings/{id}` **не используется**.

При конфликте рабочего расписания или попытке удалить интервал, содержащий подтверждённые занятия, возвращается `409` (тот же формат ошибок). Не каждая операция объявляет все возможные ошибки админки: точные коды для каждого маршрута приведены в разделе 5.

## 2.5. Админка — основные сущности

Для `PATCH` требуется минимум одно переданное изменяемое поле. **Явный `null` в nullable-поле PATCH не очищает значение:** он обрабатывается как отсутствие изменения; при передаче только таких null-полей запрос может завершиться `400`. Для текстового поля можно передать `""`, если требуется пустая строка, однако это не SQL NULL. Для необязательных полей в `POST` действуют указанные значения по умолчанию; `description`, `image_url` могут быть `null` там, где это допускает БД/DTO. Создание возвращает полный объект, обновление — полный объект. `DELETE` не возвращает JSON.

| Маршрут | JSON запроса | Ответ и правило |
|---|---|---|
| `POST /admin/services` | **`name*`, `duration_minutes*` (>0), `price*` (>=0)**, `description?`, `is_active?` (default true) | `201`: `{id,name,description,duration_minutes,price,is_active}` |
| `PATCH /admin/services/{id}` | Любое из `name`, `description`, `duration_minutes`, `price`, `is_active` | `200`: тот же полный объект |
| `DELETE /admin/services/{id}` | Без тела | `204`; деактивация `is_active=false` |
| `POST /admin/instructors` | **`user_id*`, `specialization*`, `experience_years*`**, `description?`, `image_url?`, `is_active?` (default true); `user_id` существующего пользователя с ролью `client` или `instructor`; backend сам назначает роль `instructor` | `201`: `{id,name,specialization,description,experience_years,image_url,is_active}` |
| `PATCH /admin/instructors/{id}` | Любое из `specialization`, `description`, `experience_years`, `image_url`, `is_active` | `200`: полный объект инструктора |
| `DELETE /admin/instructors/{id}` | Без тела | `204`; деактивация `is_active=false` |
| `POST /admin/equipment` | **`category_id*`, `name*`, `price_per_hour*`**, `description?`, `image_url?` | `201`: `{id,category_id,name,description,image_url,price_per_hour}` |
| `PATCH /admin/equipment/{id}` | Любое из `category_id`, `name`, `description`, `image_url`, `price_per_hour` | `200`: полный объект административного Equipment |
| `DELETE /admin/equipment/{id}` | Без тела | `204` при отсутствии связанных данных; иначе `409` |
| `GET /admin/equipment/units` | Query `equipment_id?` | `200`: `EquipmentUnit[]` |
| `POST /admin/equipment/units` | **`equipment_id*`, `inventory_number*`, `size*`, `status*`** | `201`: `EquipmentUnit` |
| `PATCH /admin/equipment/units/{id}` | Любое из `inventory_number`, `size`, `status` | `200`: `EquipmentUnit` |
| `POST /admin/equipment/sets` | **`name*`, `sport_type*`, `age_group*`, `price_per_hour*`, `items*`**, `description?`, `is_active?` | `201`: `EquipmentSetDetail` |
| `PATCH /admin/equipment/sets/{id}` | Любое из полей создания; если `items` есть — заменяет состав | `200`: `EquipmentSetDetail` |
| `DELETE /admin/equipment/sets/{id}` | Без тела | `204`; деактивация `is_active=false` |

`EquipmentUnit`: `{id:integer,equipment_id:integer,inventory_number:string,size:string,status:"available"|"maintenance"|"written_off",created_at:date-time}`.

Административный объект оборудования: `{id:integer,category_id:integer,name:string,description:string|null,image_url:string|null,price_per_hour:integer}`. Он отличается от публичного каталожного DTO, где есть `category` и `quantity`.

`EquipmentSetDetail`: `{id,name,description?,sport_type:"ski"|"snowboard",age_group:"adult"|"child",price_per_hour,is_active,items:[{category_id,category_name,quantity}]}`. В запросе комплекта `items` имеют только `category_id` и `quantity`, а в ответе также `category_name`.

Для административного чтения услуг и инструкторов теперь используются **`GET /admin/services`, `GET /admin/services/{id}`, `GET /admin/instructors`, `GET /admin/instructors/{id}`**. Публичный `/services` и `/services/{id}` показывают только активные услуги; не использовать их для чтения деактивированных услуг администратором. Для оборудования доступны публичные `GET /equipment` и `GET /equipment/{id}`, а `/admin/equipment/units` — для экземпляров. Отдельных `GET /admin/equipment`, `GET /admin/equipment/{id}` и CRUD категорий в этой версии нет.

У публичных `/instructors` и `/equipment/sets` сейчас нет фильтра только активных объектов: ответ содержит `is_active`; интерфейс должен учитывать его. Активность инструктора и услуги проверяется при создании занятия, активность комплекта — при создании аренды.

В `POST /admin/equipment/sets` поле `is_active` по умолчанию равно `true`. Состав `items` должен быть непустым; `category_id` положительный, `quantity >= 1`, повторяющиеся категории запрещены. При PATCH переданный `items` полностью заменяет состав и также должен быть непустым. Статусы экземпляров: `available`, `maintenance`, `written_off`.

Объявленные коды ошибок перечислены пооперационно в разделе 5. Некоторые операции объявляют `422`, но фактическая валидация их текущих обработчиков возвращает `400`; не считать `422` единственным кодом валидации. Удаление услуги/инструктора/комплекта означает деактивацию. Удаление модели оборудования с зависимыми строками возвращает `409`.

## 2.6. Добавленные административные операции — реализовано

Все 7 операций требуют JWT и роль `admin`: отсутствие JWT — `401`, другая роль — `403`. Все path ID — положительные целые. GET и DELETE не принимают JSON-тело.

### DTO ответов

`AdminService`: все поля обязательны; `description` может быть `null`.

```json
{
  "id": 3,
  "name": "Индивидуальное занятие",
  "description": null,
  "duration_minutes": 60,
  "price": 2000,
  "is_active": false
}
```

Типы: `id`, `duration_minutes`, `price` — integer; `name` — string; `description` — string|null; `is_active` — boolean. Поля совпадают с административными ответами создания/изменения услуги; в публичном Service обязательность `description` отличается.

`AdminInstructor`: все поля обязательны; `description` и `image_url` могут быть `null`. `id` — ID профиля инструктора, не `users.id`; поля `user_id` в этом ответе нет.

```json
{
  "id": 1,
  "name": "Иван",
  "specialization": "Лыжи",
  "description": null,
  "experience_years": 5,
  "image_url": null,
  "is_active": true
}
```

Типы: `id`, `experience_years` — integer; `name`, `specialization` — string; `description`, `image_url` — string|null; `is_active` — boolean.

### GET `/admin/services`

- Параметров и тела нет.
- `200`: массив `AdminService[]` со всеми активными и неактивными услугами, порядок по `id`; пустой результат — `[]`.
- Ошибки: `401`, `403`, `500`.

### GET `/admin/services/{id}`

- Path `id`: ID услуги. Тела нет.
- `200`: `AdminService` независимо от активности.
- Ошибки: `400`, `401`, `403`, `404`, `500`.

### GET `/admin/instructors`

- Параметров и тела нет.
- `200`: массив `AdminInstructor[]`, включая неактивных инструкторов, порядок по `id`; пустой результат — `[]`.
- Ошибки: `401`, `403`, `500`.

### GET `/admin/instructors/{id}`

- Path `id`: ID профиля инструктора. Тела нет.
- `200`: `AdminInstructor` независимо от активности.
- Ошибки: `400`, `401`, `403`, `404`, `500`.

### GET `/admin/instructors/{id}/services`

- Path `id`: ID профиля инструктора. Тела нет.
- `200`: массив назначенных `AdminService[]`, включая неактивные услуги, порядок по `id`.
- Если инструктор существует, но назначений нет — `200` и `[]`. Если инструктора нет — `404`.
- Ошибки: `400`, `401`, `403`, `404`, `500`.

### POST `/admin/instructors/{id}/services`

- Path `id`: ID профиля инструктора.
- Обязательное JSON-тело; `service_id` обязателен, integer >= 1:

```json
{"service_id": 3}
```

- `204`: услуга назначена, **тела ответа нет**. Не пытаться читать объект услуги из этого ответа.
- Назначение возможно существующему инструктору и для существующей услуги, включая неактивные объекты. Возможность бронирования отдельно зависит от активности и расписания.
- Повтор той же пары инструктор/услуга — `409`. Инструктор или услуга не существуют — `404`.
- Ошибки: `400`, `401`, `403`, `404`, `409`, `500`.

### DELETE `/admin/instructors/{id}/services/{serviceId}`

- Path `id`: ID профиля инструктора; `serviceId`: ID услуги. JSON-тела нет.
- `204`: назначение снято, **тела ответа нет**.
- Отсутствующая связь, в том числе при повторном удалении, — `404`.
- Удаляется только назначение: услуга, профиль инструктора и существующие бронирования сохраняются. Новое бронирование этой услуги у инструктора больше не проходит проверку назначения.
- Ошибки: `400`, `401`, `403`, `404`, `500`.

### Уточнение создания инструктора

`POST /admin/instructors` принимает **ID существующего пользователя**, а не данные для регистрации нового аккаунта:

```json
{
  "user_id": 3,
  "specialization": "Лыжи",
  "experience_years": 5,
  "description": null,
  "image_url": null,
  "is_active": true
}
```

Обязательны `user_id`, `specialization`, `experience_years`; `is_active` по умолчанию `true`, опыт >= 0, специализация после удаления крайних пробелов непустая. Backend в одной транзакции назначает пользователю роль `instructor` и создаёт профиль. Исходная роль может быть `client` или `instructor`; администратора преобразовать нельзя. Повторное создание профиля — `409`, неподходящий/отсутствующий пользователь — `404`.

Успех — `201`, ответ `AdminInstructor`. `name`, `email`, `password` вместо `user_id` этот метод не поддерживает. Уже выданный JWT сохраняет прежнюю роль: после назначения роли требуется повторный вход. `/auth/me` читает актуальные данные пользователя, однако это не обновляет старый JWT.

---

# 3. Что не включаем в первую версию

- `PATCH /lesson-bookings/{id}/reschedule`: перенос занятия — вне объёма сдачи.
- `POST /auth/logout`: при текущем Bearer JWT выход на фронте — удаление сохранённого токена; серверное аннулирование токена не заявляется.
- Онлайн-оплата `POST /bookings/{id}/pay`, отзывы `POST /feedback`, редактирование профиля `/me/profile` и пароль `/me/password`, аналитика `/admin/stats`, роли пользователей `/admin/users`: не описаны текущим OpenAPI и не утверждены как часть нового обязательного контракта. Не использовать как работающие API.

---

# 4. Уточнения для интеграции с текущим backend

## Маршруты и преобразования

| Сценарий / старое название | Актуальное API и формат |
|---|---|
| Каталог занятий `/lessons` | Публичные `GET /services`, `GET /services/{id}`. `name`, `duration_minutes`, `is_active`; полей `category` и `imageUrl` у услуги нет |
| Слоты `/instructors/{id}/slots?lessonId=...` | `GET /instructors/{id}/available-slots?date=YYYY-MM-DD&service_id=...`; ответ `{starts_at,ends_at}[]` |
| Создание занятия | `POST /lesson-bookings`, JSON `{instructor_id,service_id,starts_at}` |
| Создание аренды | `POST /equipment-bookings`, JSON `{starts_at,ends_at,items?,sets?}`; хотя бы одна позиция обязательна |
| Доступность `/equipment/{id}/availability` | `GET /equipment/availability?starts_at=...&ends_at=...&equipment_id=...`; массив `{equipment_id,size,available_quantity}` |
| Общий список `/me/bookings` | Объединять результаты `/lesson-bookings/my` и `/equipment-bookings/my`; ключ записи — пара тип/ID, ID из двух таблиц могут совпадать |
| Детали `/bookings/{id}` | `/lesson-bookings/{id}` или `/equipment-bookings/{id}` по типу записи |
| Отмена | `PATCH /lesson-bookings/{id}/cancel` или `PATCH /equipment-bookings/{id}/cancel`, без тела |
| Завершение занятия инструктором | `PATCH /instructors/me/bookings/{id}/status`, JSON `{status:"completed"}` либо `{status:"no_show"}` |
| Административные бронирования | Раздельные `/admin/lesson-bookings` и `/admin/equipment-bookings`; общего `/admin/bookings` нет |
| Административные услуги | Чтение и управление через `/admin/services`; `/admin/lessons` отсутствует |
| Рабочие интервалы инструктора | Административные `/admin/instructors/{id}/schedules`; самостоятельного API `/instructor/availability` нет |
| Типы frontend в camelCase | Преобразовать `name`↔`title`, `duration_minutes`↔`durationMin`, `is_active`↔`active`, `starts_at/ends_at`↔`start/end`, `total_price`↔`total` там, где это нужно модели интерфейса |

Существующие базовые API не переименованы ради старых вызовов frontend. Статусы аренды и занятий различаются; использовать enum своего типа записи. Сервер вычисляет владельца, цену, окончание занятия и исходный статус. Передача этих полей от клиента не заменяет серверный расчёт.

## Время, размеры и доступность

- Передавать date-time с часовым поясом: например `2026-11-22T10:00:00+03:00` или эквивалент в UTC с `Z`. Значение `datetime-local` без смещения не подходит для декодирования в Go `time.Time`. Выбор часового пояса ввода должен быть понятен пользователю; время расписания/слотов определяется по Москве.
- Параметры URL кодировать через URLSearchParams или аналог: знак `+` в смещении времени должен корректно экранироваться.
- Аренда начинается в будущем, длительность — минимум час и целое число часов. Это требование к длительности, а не обязательное начало ровно в HH:00.
- `size` обязателен у каждой отдельной позиции аренды. В публичном Equipment нет `sizes`: получать доступные размеры и остатки через `/equipment/availability` после выбора периода.
- `quantity` в каталоге — общее количество экземпляров, включая разные статусы, без вычета занятых на период; не использовать его вместо `available_quantity`.
- Коды категорий получать из `/equipment/categories`. Примеры `ski`, `snowboard`, `jacket`, `pants`, `boots`, `helmet`, `goggles` не являются ограничением enum в текущей схеме БД; фактический набор определяют данные. `kit` — обозначение комплектов в интерфейсе, не автоматически созданная категория API.
- Комплект — отдельная сущность. Его ID передаётся в `sets[].equipment_set_id`, а не в `items[].equipment_id`. Для каждой категории состава нужен ровно один размер. Список комплектов не содержит `items`; состав получать через `/equipment/sets/{id}`.
- Свободное время может измениться между GET и POST. Обрабатывать `409` и обновлять доступность. Шаг выдачи слотов — 30 минут от начала рабочего интервала; сервер создания записи не проверяет точное совпадение с выданным слотом.
- Время в примерах условное: при ручной проверке выбирать будущие даты и реально существующие ID/размеры.

## Обязательность, null, PATCH и ошибки

- Отсутствующее необязательное поле и явный `null` — не одно и то же по OpenAPI. Не отправлять `null` произвольно в поля, где он не разрешён. В ответах обязательность и nullability смотреть у DTO конкретной операции.
- В административных PATCH `description: null` и `image_url: null` не очищают данные: Go-поля и SQL COALESCE сохраняют старое значение. Если других изменяемых полей нет, возможен `400`. Это ограничение реализации, а не обещание удаления значения.
- При `204` ответа JSON нет. Для списков успешный пустой результат — `[]`, а не `null`.
- Типовой JSON ошибки: `{ "error": "...", "message": "..." }`, но ошибка авторизации/разбора запроса может быть текстовой. Нельзя безусловно вызывать JSON.parse на любом ошибочном ответе.
- При ошибке получения публичного списка комплектов текущий обработчик может вернуть `500` с `{}`; предусмотреть сообщение по умолчанию.
- Тексты сообщений не использовать как стабильные идентификаторы. `error` тоже не полностью унифицирован между модулями. Для занятого времени: `409`, `booking_conflict`, «Выбранное время недоступно». Для недостатка оборудования: `409`, `equipment_unavailable`. Для невозможной отмены занятия: `409`, `booking_cannot_cancel`. Для невозможной отмены аренды и смены статуса инструктором: `409`, `invalid_status_transition`. Административные конфликты: `409`, `conflict`.
- Есть известное расхождение маршрутизации: анонимный GET с нечисловым/неположительным ID для `/equipment/{id}`, `/equipment/sets/{id}`, `/instructors/{id}`, `/services/{id}` может получить `401` вместо заявленного `400`, потому что middleware не распознаёт такой путь как публичный. Это не превращает валидные публичные GET в защищённые.
- Публичные услуги выдаются только активные. Публичные инструкторы и комплекты сейчас могут содержать неактивные объекты; проверять `is_active`. Административные GET услуг/инструкторов и назначений показывают все состояния.
- После назначения пользователю роли инструктора нужен повторный вход для получения JWT с новой ролью. При выходе frontend удаляет токен локально; серверного logout/отзыва JWT нет.

## Проверки ввода и статусы

- Регистрация: `name` после trim — 2–100 символов; `email` — адрес без отображаемого имени; `phone` — `+7` и 10 цифр; пароль — 8–72 байта. Регистр email автоматически не нормализуется. Зарегистрированный пользователь получает роль `client`.
- При создании/изменении услуги: непустое имя, `duration_minutes > 0`, `price >= 0`. Цены — целые рубли, не копейки.
- Рабочий интервал: `ends_at > starts_at`; интервалы одного инструктора не должны пересекаться. Нельзя удалить/сократить интервал так, чтобы оставить уже видимые подтверждённые занятия вне него. Даты фильтров расписания/административных броней трактуются по Москве.
- Занятия: `confirmed → cancelled` только до начала; `confirmed → completed/no_show` после начала. Финальные статусы не переключаются обратно. Попытка передать иной целевой статус административному PATCH не означает произвольное присвоение статуса и может вернуть `400`; запрещённый переход из текущего состояния — `409`.
- Аренда: `pending → confirmed/cancelled`, `confirmed → issued`, `confirmed → cancelled` только до начала, `issued → returned`. Клиент может отменить `pending/confirmed` только до начала. Новая клиентская аренда сразу `confirmed`; отдельной оплаты нет. Backend не привязывает переход `issued/returned` к дополнительной проверке часа выдачи/возврата.

## Границы подтверждённого поведения

Защита от пересечения занятий одного инструктора предусмотрена ограничением PostgreSQL; аренда распределяет экземпляры в транзакции с общей advisory-блокировкой. Эти механизмы не означают, что все конкурентные административные изменения уже проверены: создание занятия и проверка расписания/активности/назначения выполняются отдельными запросами, поэтому остаются окна при одновременном изменении этих данных. Запрет пересекающихся занятий одного клиента у разных инструкторов отдельно не реализован. Сквозные и конкурентные проверки на реальной БД для всего API ещё не подтверждены.

Этот документ описывает существующий backend; упомянутые ограничения здесь не исправлены. Frontend-код и данные БД при обновлении документа не менялись.

---

# 5. Полный реестр реализованных операций и HTTP-кодов

Реестр сверён с текущим OpenAPI. `2xx` — объявленный успешный ответ; ошибки перечислены по схеме конкретной операции. Наличие кода в схеме не означает, что каждый путь ошибки был воспроизведён на БД. Текстовые ошибки middleware и известные особенности из раздела 4 сохраняются.

| Метод | Маршрут | Успех | Ошибки по OpenAPI |
|---|---|---|---|
| POST | `/auth/register` | 201 | 400, 409, 500 |
| POST | `/auth/login` | 200 | 400, 401, 500 |
| GET | `/auth/me` | 200 | 401, 500 |
| GET | `/equipment` | 200 | 500 |
| GET | `/equipment/{id}` | 200 | 400, 404, 500 |
| GET | `/equipment/categories` | 200 | 500 |
| GET | `/equipment/sets` | 200 | 500 |
| GET | `/equipment/sets/{id}` | 200 | 400, 404, 500 |
| GET | `/instructors` | 200 | 500 |
| GET | `/instructors/{id}` | 200 | 404, 400, 500 |
| GET | `/instructors/{id}/schedule` | 200 | 404, 400, 500 |
| GET | `/instructors/{id}/available-slots` | 200 | 400, 404, 422, 500 |
| GET | `/instructors/me/bookings` | 200 | 401, 400, 403, 500 |
| GET | `/instructors/{id}/services` | 200 | 400, 404, 500 |
| PATCH | `/instructors/me/bookings/{id}/status` | 200 | 400, 401, 403, 404, 409, 500 |
| GET | `/services` | 200 | 500 |
| GET | `/services/{id}` | 200 | 400, 404, 500 |
| POST | `/lesson-bookings` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/lesson-bookings/my` | 200 | 400, 401, 403, 500 |
| GET | `/lesson-bookings/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/lesson-bookings/{id}/cancel` | 200 | 400, 401, 403, 404, 409, 500 |
| GET | `/equipment/availability` | 200 | 400, 404, 500 |
| POST | `/equipment-bookings` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/equipment-bookings/my` | 200 | 400, 401, 403, 500 |
| GET | `/equipment-bookings/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/equipment-bookings/{id}/cancel` | 200 | 400, 401, 403, 404, 409, 500 |
| GET | `/admin/services` | 200 | 401, 403, 500 |
| POST | `/admin/services` | 201 | 400, 401, 403, 409, 422, 500 |
| GET | `/admin/services/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/admin/services/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| DELETE | `/admin/services/{id}` | 204 | 400, 401, 403, 404, 409, 422, 500 |
| POST | `/admin/instructors` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/admin/instructors` | 200 | 401, 403, 500 |
| GET | `/admin/instructors/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/admin/instructors/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| DELETE | `/admin/instructors/{id}` | 204 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/admin/instructors/{id}/services` | 200 | 400, 401, 403, 404, 500 |
| POST | `/admin/instructors/{id}/services` | 204 | 400, 401, 403, 404, 409, 500 |
| DELETE | `/admin/instructors/{id}/services/{serviceId}` | 204 | 400, 401, 403, 404, 500 |
| GET | `/admin/instructors/{id}/schedules` | 200 | 400, 401, 403, 404, 500 |
| POST | `/admin/instructors/{id}/schedules` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| PATCH | `/admin/instructors/{id}/schedules/{scheduleId}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| DELETE | `/admin/instructors/{id}/schedules/{scheduleId}` | 204 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/admin/lesson-bookings` | 200 | 400, 401, 403, 500 |
| GET | `/admin/lesson-bookings/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/admin/lesson-bookings/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/admin/equipment-bookings` | 200 | 400, 401, 403, 500 |
| GET | `/admin/equipment-bookings/{id}` | 200 | 400, 401, 403, 404, 500 |
| PATCH | `/admin/equipment-bookings/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| POST | `/admin/equipment` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| PATCH | `/admin/equipment/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| DELETE | `/admin/equipment/{id}` | 204 | 400, 401, 403, 404, 409, 422, 500 |
| GET | `/admin/equipment/units` | 200 | 400, 401, 403, 500 |
| POST | `/admin/equipment/units` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| PATCH | `/admin/equipment/units/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| POST | `/admin/equipment/sets` | 201 | 400, 401, 403, 404, 409, 422, 500 |
| PATCH | `/admin/equipment/sets/{id}` | 200 | 400, 401, 403, 404, 409, 422, 500 |
| DELETE | `/admin/equipment/sets/{id}` | 204 | 400, 401, 403, 404, 409, 422, 500 |
