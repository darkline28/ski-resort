/**
 * Заглушка бэкенда: эмулирует REST API из docs/API.md поверх localStorage.
 * Логика (свободные интервалы, двойное бронирование, права) здесь только для демо —
 * в продакшене её выполняет сервер. Фронт обращается к ней через тот же request().
 */
import type { Booking, BookingStatus, Role, ScheduleCell, Slot, User } from '@/types';
import { getDb, nextId, saveDb, resetDb, type MockUser } from './db';
import { startOfDay } from '@/utils/date';

interface MockResponse { status: number; body: unknown }

class HttpError extends Error {
  constructor(public status: number, message: string, public errors?: Record<string, string>) {
    super(message);
  }
}

type Ctx = {
  params: Record<string, string>;
  query: URLSearchParams;
  body: any; // eslint-disable-line @typescript-eslint/no-explicit-any
  user: MockUser | null;
};
type Handler = (ctx: Ctx) => unknown;

const routes: { method: string; pattern: RegExp; keys: string[]; handler: Handler }[] = [];

function route(method: string, path: string, handler: Handler) {
  const keys: string[] = [];
  const pattern = new RegExp(
    '^' + path.replace(/:(\w+)/g, (_, k) => { keys.push(k); return '([^/]+)'; }) + '$',
  );
  routes.push({ method, pattern, keys, handler });
}

const delay = (ms: number) => new Promise((r) => setTimeout(r, ms));

export async function mockFetch(method: string, url: string, body: unknown, token: string | null): Promise<MockResponse> {
  await delay(250 + Math.random() * 300);
  const [path, qs] = url.split('?');
  const query = new URLSearchParams(qs);
  const db = getDb();
  const userId = token?.startsWith('mock-token-') ? Number(token.slice('mock-token-'.length)) : null;
  const user = userId ? db.users.find((u) => u.id === userId) ?? null : null;

  for (const r of routes) {
    if (r.method !== method) continue;
    const m = path.match(r.pattern);
    if (!m) continue;
    const params: Record<string, string> = {};
    r.keys.forEach((k, i) => { params[k] = decodeURIComponent(m[i + 1]); });
    try {
      const result = r.handler({ params, query, body: body ?? {}, user });
      saveDb();
      return { status: method === 'POST' && result !== undefined ? 201 : result === undefined ? 204 : 200, body: clone(result) };
    } catch (e) {
      if (e instanceof HttpError) return { status: e.status, body: { message: e.message, errors: e.errors } };
      console.error('[mock]', e);
      return { status: 500, body: { message: 'Внутренняя ошибка сервера' } };
    }
  }
  return { status: 404, body: { message: `Маршрут ${method} ${path} не найден` } };
}

const clone = <T>(v: T): T => (v === undefined ? v : JSON.parse(JSON.stringify(v)));

// ---------- helpers ----------
const publicUser = ({ password: _p, ...u }: MockUser): User => u;

function auth(ctx: Ctx, ...roles: Role[]): MockUser {
  if (!ctx.user) throw new HttpError(401, 'Нужно войти в аккаунт.');
  if (roles.length && !roles.includes(ctx.user.role)) throw new HttpError(403, 'Недостаточно прав.');
  return ctx.user;
}

const isEmail = (s: unknown) => typeof s === 'string' && /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s);

function requireFields(fields: Record<string, unknown>) {
  const errors: Record<string, string> = {};
  Object.entries(fields).forEach(([k, v]) => {
    if (v === undefined || v === null || String(v).trim() === '') errors[k] = 'Обязательное поле';
  });
  if (Object.keys(errors).length) throw new HttpError(422, 'Проверь правильность заполнения полей.', errors);
}

const overlaps = (aS: string, aE: string, bS: string, bE: string) =>
  new Date(aS) < new Date(bE) && new Date(bS) < new Date(aE);

const ACTIVE: BookingStatus[] = ['pending', 'paid'];

function instructorByUser(u: MockUser) {
  const ins = getDb().instructors.find((i) => i.userId === u.id);
  if (!ins) throw new HttpError(404, 'Профиль инструктора не найден.');
  return ins;
}

/** Свободен ли интервал у инструктора: покрыт рабочими окнами и не пересекается с бронями */
function isInstructorFree(instructorId: number, start: string, end: string, ignoreBookingId?: number) {
  const db = getDb();
  const s = new Date(start).getTime();
  const e = new Date(end).getTime();
  for (let t = s; t < e; t += 3600000) {
    const covered = db.availability.some(
      (a) => a.instructorId === instructorId && new Date(a.start).getTime() <= t && new Date(a.end).getTime() > t,
    );
    if (!covered) return false;
  }
  return !db.bookings.some(
    (b) => b.id !== ignoreBookingId && b.instructorId === instructorId && ACTIVE.includes(b.status) &&
      overlaps(b.start, b.end, start, end),
  );
}

function rentedCount(equipmentId: number, size: string, start: string, end: string) {
  return getDb().bookings
    .filter((b) => b.equipmentId === equipmentId && b.size === size && ACTIVE.includes(b.status) && overlaps(b.start, b.end, start, end))
    .reduce((sum, b) => sum + (b.quantity ?? 1), 0);
}

function findBooking(id: string) {
  const b = getDb().bookings.find((x) => x.id === Number(id));
  if (!b) throw new HttpError(404, 'Бронирование не найдено.');
  return b;
}

const sortByStart = (a: Booking, b: Booking) => new Date(a.start).getTime() - new Date(b.start).getTime();

// ---------- auth ----------
route('POST', '/auth/login', ({ body }) => {
  requireFields({ email: body.email, password: body.password });
  const u = getDb().users.find((x) => x.email.toLowerCase() === String(body.email).toLowerCase());
  if (!u || u.password !== body.password) throw new HttpError(401, 'Неверный email или пароль.');
  return { token: `mock-token-${u.id}`, user: publicUser(u) };
});

route('POST', '/auth/register', ({ body }) => {
  requireFields({ name: body.name, phone: body.phone, email: body.email, password: body.password });
  if (!isEmail(body.email)) throw new HttpError(422, 'Некорректный email', { email: 'Введите корректный email' });
  if (String(body.password).length < 6) throw new HttpError(422, 'Слабый пароль', { password: 'Пароль — минимум 6 символов' });
  const db = getDb();
  if (db.users.some((u) => u.email.toLowerCase() === String(body.email).toLowerCase())) {
    throw new HttpError(409, 'Пользователь с таким email уже существует.', { email: 'Email уже зарегистрирован' });
  }
  const u: MockUser = { id: nextId(), name: body.name, phone: body.phone, email: body.email, role: 'client', password: body.password };
  db.users.push(u);
  return { token: `mock-token-${u.id}`, user: publicUser(u) };
});

route('POST', '/auth/logout', () => undefined);
route('GET', '/auth/me', (ctx) => publicUser(auth(ctx)));

// ---------- public catalog ----------
route('GET', '/lessons', ({ query }) => {
  const cat = query.get('category');
  return getDb().lessons.filter((l) => l.active && (!cat || l.category === cat));
});
route('GET', '/lessons/:id', ({ params }) => {
  const l = getDb().lessons.find((x) => x.id === Number(params.id) && x.active);
  if (!l) throw new HttpError(404, 'Занятие не найдено.');
  return l;
});
route('GET', '/equipment', ({ query }) => {
  const cat = query.get('category');
  return getDb().equipment.filter((e) => e.active && (!cat || e.category === cat));
});
route('GET', '/equipment/:id', ({ params }) => {
  const e = getDb().equipment.find((x) => x.id === Number(params.id) && x.active);
  if (!e) throw new HttpError(404, 'Снаряжение не найдено.');
  return e;
});
route('GET', '/equipment/:id/availability', ({ params, query }) => {
  const e = getDb().equipment.find((x) => x.id === Number(params.id));
  if (!e) throw new HttpError(404, 'Снаряжение не найдено.');
  const start = query.get('start')!, end = query.get('end')!, size = query.get('size')!;
  return { available: Math.max(0, e.stock - rentedCount(e.id, size, start, end)) };
});
route('GET', '/instructors', ({ query }) => {
  const db = getDb();
  const lessonId = Number(query.get('lessonId'));
  const lesson = lessonId ? db.lessons.find((l) => l.id === lessonId) : null;
  return db.instructors.filter((i) => !lesson || i.categories.includes(lesson.category));
});
route('GET', '/instructors/:id', ({ params }) => {
  const i = getDb().instructors.find((x) => x.id === Number(params.id));
  if (!i) throw new HttpError(404, 'Инструктор не найден.');
  return i;
});

/** Свободные интервалы — считаются «на сервере» */
route('GET', '/instructors/:id/slots', ({ params, query }) => {
  const db = getDb();
  const instructorId = Number(params.id);
  const date = query.get('date');
  if (!date) throw new HttpError(400, 'Не указана дата.');
  const lesson = db.lessons.find((l) => l.id === Number(query.get('lessonId')));
  const duration = (lesson?.durationMin ?? 60) * 60000;
  const [y, m, d] = date.split('-').map(Number);
  const dayStart = new Date(y, m - 1, d).getTime();
  const dayEnd = dayStart + 86400000;
  const now = Date.now();

  const slots: Slot[] = db.availability
    .filter((a) => a.instructorId === instructorId && new Date(a.start).getTime() >= dayStart && new Date(a.start).getTime() < dayEnd)
    .sort((a, b) => a.start.localeCompare(b.start))
    .map((a) => {
      const start = a.start;
      const end = new Date(new Date(start).getTime() + duration).toISOString();
      const available = new Date(start).getTime() > now && isInstructorFree(instructorId, start, end);
      return { start, end, available };
    });
  return slots;
});

route('POST', '/feedback', ({ body }) => {
  requireFields({ name: body.name, contact: body.contact, message: body.message });
  getDb().feedback.push({ ...body, createdAt: new Date().toISOString() });
  return { ok: true };
});

// ---------- bookings (client) ----------
route('POST', '/bookings/lesson', (ctx) => {
  const u = auth(ctx, 'client');
  const { lessonId, instructorId, start } = ctx.body;
  requireFields({ lessonId, instructorId, start });
  const db = getDb();
  const lesson = db.lessons.find((l) => l.id === lessonId && l.active);
  const ins = db.instructors.find((i) => i.id === instructorId);
  if (!lesson || !ins) throw new HttpError(404, 'Занятие или инструктор не найдены.');
  if (!ins.categories.includes(lesson.category)) throw new HttpError(422, 'Инструктор не ведёт это занятие.');
  if (new Date(start).getTime() <= Date.now()) throw new HttpError(422, 'Нельзя записаться на прошедшее время.');
  const end = new Date(new Date(start).getTime() + lesson.durationMin * 60000).toISOString();
  if (!isInstructorFree(ins.id, start, end)) {
    throw new HttpError(409, 'Это время уже занято. Выбери другой интервал.');
  }
  const b: Booking = {
    id: nextId(), type: 'lesson', status: 'pending', userId: u.id, clientName: u.name,
    title: lesson.title, lessonId, instructorId, instructorName: ins.name,
    start, end, total: lesson.price, createdAt: new Date().toISOString(),
  };
  db.bookings.push(b);
  return b;
});

route('POST', '/bookings/rental', (ctx) => {
  const u = auth(ctx, 'client');
  const { equipmentId, size, quantity, start, end } = ctx.body;
  requireFields({ equipmentId, size, quantity, start, end });
  const db = getDb();
  const e = db.equipment.find((x) => x.id === equipmentId && x.active);
  if (!e) throw new HttpError(404, 'Снаряжение не найдено.');
  if (!e.sizes.includes(size)) throw new HttpError(422, 'Нет такого размера.', { size: 'Выбери размер' });
  const hours = (new Date(end).getTime() - new Date(start).getTime()) / 3600000;
  if (hours < 1) throw new HttpError(422, 'Минимальный период аренды — 1 час.', { end: 'Минимум 1 час' });
  if (new Date(start).getTime() < Date.now()) throw new HttpError(422, 'Начало аренды уже прошло.', { start: 'Выбери будущее время' });
  const free = e.stock - rentedCount(e.id, size, start, end);
  if (quantity > free) throw new HttpError(409, `На этот период доступно только ${free} шт.`);
  const b: Booking = {
    id: nextId(), type: 'rental', status: 'pending', userId: u.id, clientName: u.name,
    title: `Аренда: ${e.title.toLowerCase()}`, equipmentId, size, quantity, start, end,
    total: Math.ceil(hours) * e.pricePerHour * quantity, createdAt: new Date().toISOString(),
  };
  db.bookings.push(b);
  return b;
});

route('GET', '/me/bookings', (ctx) => {
  const u = auth(ctx);
  const type = ctx.query.get('type');
  return getDb().bookings.filter((b) => b.userId === u.id && (!type || b.type === type)).sort(sortByStart);
});

route('GET', '/bookings/:id', (ctx) => {
  const u = auth(ctx);
  const b = findBooking(ctx.params.id);
  if (u.role === 'client' && b.userId !== u.id) throw new HttpError(403, 'Это чужое бронирование.');
  return b;
});

route('POST', '/bookings/:id/cancel', (ctx) => {
  const u = auth(ctx);
  const b = findBooking(ctx.params.id);
  if (u.role === 'client' && b.userId !== u.id) throw new HttpError(403, 'Это чужое бронирование.');
  if (!ACTIVE.includes(b.status)) throw new HttpError(422, 'Это бронирование нельзя отменить.');
  if (new Date(b.start).getTime() <= Date.now()) throw new HttpError(422, 'Отменить можно только до начала занятия.');
  b.status = 'cancelled';
  return b;
});

route('POST', '/bookings/:id/pay', (ctx) => {
  const u = auth(ctx);
  const b = findBooking(ctx.params.id);
  if (b.userId !== u.id) throw new HttpError(403, 'Это чужое бронирование.');
  if (b.status !== 'pending') throw new HttpError(422, 'Заказ уже оплачен или отменён.');
  const card = String(ctx.body.cardNumber ?? '').replace(/\s/g, '');
  if (!/^\d{16}$/.test(card)) throw new HttpError(422, 'Некорректный номер карты.', { cardNumber: '16 цифр' });
  // Демо: карта, оканчивающаяся на 0000, «отклоняется», чтобы показать экран ошибки
  if (card.endsWith('0000')) throw new HttpError(402, 'Платёж отклонён банком (демо).');
  b.status = 'paid';
  return b;
});

route('POST', '/bookings/:id/complete', (ctx) => {
  const u = auth(ctx, 'instructor', 'admin');
  const b = findBooking(ctx.params.id);
  if (u.role === 'instructor' && b.instructorId !== instructorByUser(u).id) throw new HttpError(403, 'Это не ваше занятие.');
  if (b.status !== 'paid') throw new HttpError(422, 'Завершить можно только оплаченное занятие.');
  b.status = 'completed';
  return b;
});

// ---------- profile ----------
route('PATCH', '/me/profile', (ctx) => {
  const u = auth(ctx);
  const { name, phone, email } = ctx.body;
  requireFields({ name, phone, email });
  if (!isEmail(email)) throw new HttpError(422, 'Некорректный email', { email: 'Введите корректный email' });
  if (getDb().users.some((x) => x.id !== u.id && x.email.toLowerCase() === email.toLowerCase())) {
    throw new HttpError(409, 'Email уже занят.', { email: 'Email уже занят' });
  }
  Object.assign(u, { name, phone, email });
  return publicUser(u);
});

route('POST', '/me/password', (ctx) => {
  const u = auth(ctx);
  if (ctx.body.currentPassword !== u.password) throw new HttpError(422, 'Текущий пароль неверный.', { currentPassword: 'Неверный пароль' });
  if (String(ctx.body.newPassword ?? '').length < 6) throw new HttpError(422, 'Слабый пароль', { newPassword: 'Минимум 6 символов' });
  u.password = ctx.body.newPassword;
  return undefined;
});

// ---------- instructor ----------
route('GET', '/instructor/me', (ctx) => instructorByUser(auth(ctx, 'instructor')));

route('GET', '/instructor/bookings', (ctx) => {
  const ins = instructorByUser(auth(ctx, 'instructor'));
  const from = ctx.query.get('from')!, to = ctx.query.get('to')!;
  return getDb().bookings
    .filter((b) => b.instructorId === ins.id && b.status !== 'cancelled' && overlaps(b.start, b.end, from, to))
    .sort(sortByStart);
});

route('GET', '/instructor/schedule', (ctx) => {
  const ins = instructorByUser(auth(ctx, 'instructor'));
  const from = ctx.query.get('from')!, to = ctx.query.get('to')!;
  const db = getDb();
  const cells: ScheduleCell[] = db.availability
    .filter((a) => a.instructorId === ins.id && overlaps(a.start, a.end, from, to))
    .map((a) => {
      const b = db.bookings.find((x) => x.instructorId === ins.id && ACTIVE.concat('completed').includes(x.status) && overlaps(x.start, x.end, a.start, a.end));
      return { start: a.start, end: a.end, status: b ? 'booked' : 'free', availabilityId: a.id, bookingId: b?.id };
    });
  return cells;
});

route('POST', '/instructor/availability', (ctx) => {
  const ins = instructorByUser(auth(ctx, 'instructor'));
  const { start, end } = ctx.body;
  requireFields({ start, end });
  if (new Date(end) <= new Date(start)) throw new HttpError(422, 'Окончание должно быть позже начала.', { end: 'Позже начала' });
  const db = getDb();
  // Режем на часовые окна, пропуская уже существующие
  const created = [];
  for (let t = new Date(start).getTime(); t < new Date(end).getTime(); t += 3600000) {
    const s = new Date(t).toISOString();
    const e = new Date(t + 3600000).toISOString();
    if (db.availability.some((a) => a.instructorId === ins.id && overlaps(a.start, a.end, s, e))) continue;
    const a = { id: nextId(), instructorId: ins.id, start: s, end: e };
    db.availability.push(a);
    created.push(a);
  }
  if (!created.length) throw new HttpError(409, 'Это время уже отмечено как рабочее.');
  return created[0];
});

route('DELETE', '/instructor/availability/:id', (ctx) => {
  const ins = instructorByUser(auth(ctx, 'instructor'));
  const db = getDb();
  const a = db.availability.find((x) => x.id === Number(ctx.params.id) && x.instructorId === ins.id);
  if (!a) throw new HttpError(404, 'Окно не найдено.');
  if (db.bookings.some((b) => b.instructorId === ins.id && ACTIVE.includes(b.status) && overlaps(b.start, b.end, a.start, a.end))) {
    throw new HttpError(409, 'На это время уже есть запись клиента.');
  }
  db.availability = db.availability.filter((x) => x !== a);
  return undefined;
});

// ---------- admin ----------
route('GET', '/admin/stats', (ctx) => {
  auth(ctx, 'admin');
  const db = getDb();
  const today = startOfDay(new Date());
  const tomorrow = new Date(today.getTime() + 86400000);
  const weekAgo = new Date(today.getTime() - 6 * 86400000);
  const byWeekday = [0, 0, 0, 0, 0, 0, 0];
  db.bookings.filter((b) => b.status !== 'cancelled').forEach((b) => { byWeekday[(new Date(b.start).getDay() + 6) % 7]++; });
  return {
    users: db.users.filter((u) => u.role === 'client').length,
    instructors: db.instructors.length,
    bookingsToday: db.bookings.filter((b) => b.status !== 'cancelled' && new Date(b.start) >= today && new Date(b.start) < tomorrow).length,
    revenueWeek: db.bookings
      .filter((b) => (b.status === 'paid' || b.status === 'completed') && new Date(b.start) >= weekAgo && new Date(b.start) < tomorrow)
      .reduce((s, b) => s + b.total, 0),
    bookingsByWeekday: byWeekday,
    latestBookings: [...db.bookings].sort((a, b) => b.id - a.id).slice(0, 4),
  };
});

route('GET', '/admin/users', (ctx) => {
  auth(ctx, 'admin');
  const role = ctx.query.get('role');
  return getDb().users.filter((u) => !role || u.role === role).map(publicUser);
});

route('PATCH', '/admin/users/:id', (ctx) => {
  auth(ctx, 'admin');
  const u = getDb().users.find((x) => x.id === Number(ctx.params.id));
  if (!u) throw new HttpError(404, 'Пользователь не найден.');
  const { name, phone, email } = ctx.body;
  if (email !== undefined && !isEmail(email)) throw new HttpError(422, 'Некорректный email', { email: 'Введите корректный email' });
  Object.assign(u, Object.fromEntries(Object.entries({ name, phone, email }).filter(([, v]) => v !== undefined)));
  const ins = getDb().instructors.find((i) => i.userId === u.id);
  if (ins) Object.assign(ins, { name: u.name, email: u.email });
  return publicUser(u);
});

route('POST', '/admin/instructors', (ctx) => {
  auth(ctx, 'admin');
  const { name, email, password, specialization } = ctx.body;
  requireFields({ name, email, password, specialization });
  if (!isEmail(email)) throw new HttpError(422, 'Некорректный email', { email: 'Введите корректный email' });
  if (String(password).length < 6) throw new HttpError(422, 'Слабый пароль', { password: 'Минимум 6 символов' });
  const db = getDb();
  if (db.users.some((u) => u.email.toLowerCase() === email.toLowerCase())) throw new HttpError(409, 'Email уже занят.', { email: 'Email уже занят' });
  const u: MockUser = { id: nextId(), name, email, phone: '', role: 'instructor', password };
  db.users.push(u);
  const ins = { id: nextId(), userId: u.id, name, email, specialization, categories: ['ski' as const], experienceYears: 0, pricePerHour: 3000, bio: '' };
  db.instructors.push(ins);
  return ins;
});

function crud<T extends { id: number }>(path: string, list: () => T[], setList: (v: T[]) => void, label: string) {
  route('GET', `/admin/${path}`, (ctx) => { auth(ctx, 'admin'); return list(); });
  route('POST', `/admin/${path}`, (ctx) => {
    auth(ctx, 'admin');
    requireFields({ title: ctx.body.title });
    const item = { ...ctx.body, id: nextId() } as T;
    list().push(item);
    return item;
  });
  route('PATCH', `/admin/${path}/:id`, (ctx) => {
    auth(ctx, 'admin');
    const item = list().find((x) => x.id === Number(ctx.params.id));
    if (!item) throw new HttpError(404, `${label} не найдено.`);
    Object.assign(item, ctx.body, { id: item.id });
    return item;
  });
  route('DELETE', `/admin/${path}/:id`, (ctx) => {
    auth(ctx, 'admin');
    setList(list().filter((x) => x.id !== Number(ctx.params.id)));
    return undefined;
  });
}
crud('lessons', () => getDb().lessons, (v) => { getDb().lessons = v; }, 'Занятие');
crud('equipment', () => getDb().equipment, (v) => { getDb().equipment = v; }, 'Снаряжение');

route('GET', '/admin/bookings', (ctx) => {
  auth(ctx, 'admin');
  const status = ctx.query.get('status');
  return getDb().bookings.filter((b) => !status || b.status === status).sort((a, b) => b.id - a.id);
});

route('PATCH', '/admin/bookings/:id', (ctx) => {
  auth(ctx, 'admin');
  const b = findBooking(ctx.params.id);
  const status = ctx.body.status as BookingStatus;
  if (!['pending', 'paid', 'cancelled', 'completed'].includes(status)) throw new HttpError(422, 'Неизвестный статус.');
  if (ACTIVE.includes(status) && b.instructorId && !ACTIVE.includes(b.status) && !isInstructorFree(b.instructorId, b.start, b.end, b.id)) {
    throw new HttpError(409, 'Интервал уже занят другой записью.');
  }
  b.status = status;
  return b;
});

// Утилита для отладки из консоли: window.__montisResetMock()
(window as unknown as { __montisResetMock: () => void }).__montisResetMock = () => { resetDb(); location.reload(); };
