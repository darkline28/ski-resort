/**
 * In-memory «база» заглушки. Хранится в localStorage, чтобы данные переживали перезагрузку.
 * Сбросить: localStorage.removeItem('montis_mock_db') или кнопка в README.
 * В реальном проекте всё это живёт на бэкенде — фронт видит только API.
 */
import type { Availability, Booking, Equipment, Instructor, Lesson, User } from '@/types';
import { addDays, startOfDay } from '@/utils/date';

export interface MockUser extends User { password: string }

export interface MockDb {
  version: number;
  seq: number;
  users: MockUser[];
  lessons: Lesson[];
  equipment: Equipment[];
  instructors: Instructor[];
  availability: Availability[];
  bookings: Booking[];
  feedback: { name: string; contact: string; message: string; createdAt: string }[];
}

const KEY = 'montis_mock_db';
const VERSION = 3;

export const WORK_HOURS = [9, 10, 11, 12, 14, 15, 16, 17];

function at(day: Date, hour: number, min = 0): string {
  const d = new Date(day);
  d.setHours(hour, min, 0, 0);
  return d.toISOString();
}

// Детерминированный «рандом», чтобы расписание выглядело живым, но стабильным
function pseudo(n: number): number {
  const x = Math.sin(n * 9301 + 49297) * 233280;
  return x - Math.floor(x);
}

function seed(): MockDb {
  const today = startOfDay(new Date());
  const now = new Date().toISOString();

  const users: MockUser[] = [
    { id: 1, name: 'Администратор', phone: '+7 900 000-00-01', email: 'admin@montis.ru', role: 'admin', password: 'admin123' },
    { id: 2, name: 'Анна Смирнова', phone: '+7 900 000-00-00', email: 'anna@example.com', role: 'client', password: '123456' },
    { id: 3, name: 'Алексей Кузнецов', phone: '+7 900 111-11-11', email: 'alexey@montis.ru', role: 'instructor', password: '123456' },
    { id: 4, name: 'Мария Соколова', phone: '+7 900 222-22-22', email: 'maria@montis.ru', role: 'instructor', password: '123456' },
    { id: 5, name: 'Иван Петров', phone: '+7 900 333-33-33', email: 'ivan@montis.ru', role: 'instructor', password: '123456' },
    { id: 6, name: 'Игорь Волков', phone: '+7 900 444-44-44', email: 'igor@example.com', role: 'client', password: '123456' },
    { id: 7, name: 'Ольга Миронова', phone: '+7 900 555-55-55', email: 'olga@example.com', role: 'client', password: '123456' },
    { id: 8, name: 'Пётр Лебедев', phone: '+7 900 666-66-66', email: 'petr@example.com', role: 'client', password: '123456' },
  ];

  const lessons: Lesson[] = [
    { id: 1, title: 'Горные лыжи', description: 'Освой технику катания и почувствуй свободу на склонах.', category: 'ski', price: 2500, durationMin: 60, active: true },
    { id: 2, title: 'Сноуборд', description: 'Попробуй себя в сноубординге вместе с опытным инструктором.', category: 'snowboard', price: 2800, durationMin: 60, active: true },
    { id: 3, title: 'Индивидуальное обучение', description: 'Персональный подход и программа занятий под твой уровень.', category: 'individual', price: 4500, durationMin: 90, active: true },
    { id: 4, title: 'Групповое занятие', description: 'Небольшая группа и дружелюбная атмосфера.', category: 'group', price: 2000, durationMin: 60, active: true },
    { id: 5, title: 'Детские занятия', description: 'Безопасное обучение для самых юных гостей.', category: 'kids', price: 2200, durationMin: 60, active: true },
    { id: 6, title: 'Фристайл', description: 'Трюки, прыжки и катание в парке.', category: 'freestyle', price: 3800, durationMin: 60, active: true },
  ];

  const equipment: Equipment[] = [
    { id: 1, title: 'Горные лыжи', description: 'Лыжи под твой рост и уровень подготовки', category: 'ski', pricePerHour: 600, sizes: ['150', '160', '170', '180'], stock: 20, active: true },
    { id: 2, title: 'Сноуборд', description: 'Доски разных размеров и жёсткости', category: 'snowboard', pricePerHour: 700, sizes: ['140', '150', '155', '160'], stock: 15, active: true },
    { id: 3, title: 'Комплект снаряжения', description: 'Лыжи или сноуборд, ботинки и шлем', category: 'kit', pricePerHour: 1400, sizes: ['36', '38', '40', '42', '44'], stock: 12, active: true },
    { id: 4, title: 'Горнолыжные ботинки', description: 'Удобные ботинки нужного размера', category: 'boots', pricePerHour: 300, sizes: ['36', '38', '40', '42', '44', '46'], stock: 30, active: true },
    { id: 5, title: 'Шлем', description: 'Защита для взрослых и детей', category: 'helmet', pricePerHour: 200, sizes: ['S', 'M', 'L', 'XL'], stock: 30, active: true },
    { id: 6, title: 'Маска', description: 'Защита глаз от солнца и снега', category: 'goggles', pricePerHour: 150, sizes: ['Детская', 'Взрослая'], stock: 25, active: true },
  ];

  const instructors: Instructor[] = [
    { id: 1, userId: 3, name: 'Алексей Кузнецов', email: 'alexey@montis.ru', specialization: 'Горные лыжи', categories: ['ski', 'individual', 'group'], experienceYears: 8, pricePerHour: 3000, bio: 'Помогаю новичкам уверенно встать на лыжи и улучшить технику.' },
    { id: 2, userId: 4, name: 'Мария Соколова', email: 'maria@montis.ru', specialization: 'Сноуборд', categories: ['snowboard', 'freestyle', 'individual'], experienceYears: 6, pricePerHour: 3200, bio: 'Сноуборд и фристайл: от первых поворотов до трюков в парке.' },
    { id: 3, userId: 5, name: 'Иван Петров', email: 'ivan@montis.ru', specialization: 'Дети и новички', categories: ['kids', 'ski', 'group'], experienceYears: 5, pricePerHour: 2800, bio: 'Терпеливо и безопасно учу детей и тех, кто впервые на склоне.' },
  ];

  let seq = 1000;
  const availability: Availability[] = [];
  for (const ins of instructors) {
    for (let d = -7; d <= 21; d++) {
      const day = addDays(today, d);
      WORK_HOURS.forEach((h, i) => {
        if (pseudo(ins.id * 1000 + d * 10 + i) < 0.18) return; // не работает в этот час
        availability.push({ id: ++seq, instructorId: ins.id, start: at(day, h), end: at(day, h + 1) });
      });
    }
  }

  const mk = (
    b: Omit<Booking, 'id' | 'createdAt' | 'clientName'> & { clientName?: string },
  ): Booking => ({
    id: ++seq,
    createdAt: now,
    clientName: b.clientName ?? users.find((u) => u.id === b.userId)!.name,
    ...b,
  });

  const lessonB = (
    userId: number, lessonId: number, instructorId: number, dayOffset: number, hour: number,
    status: Booking['status'],
  ) => {
    const l = lessons.find((x) => x.id === lessonId)!;
    const ins = instructors.find((x) => x.id === instructorId)!;
    const day = addDays(today, dayOffset);
    const start = at(day, hour);
    const end = new Date(new Date(start).getTime() + l.durationMin * 60000).toISOString();
    // гарантируем, что под бронью есть «рабочее» окно
    for (let h = hour; h < hour + Math.ceil(l.durationMin / 60); h++) {
      const s = at(day, h);
      if (!availability.some((a) => a.instructorId === instructorId && a.start === s)) {
        availability.push({ id: ++seq, instructorId, start: s, end: at(day, h + 1) });
      }
    }
    return mk({ type: 'lesson', status, userId, title: l.title, lessonId, instructorId, instructorName: ins.name, start, end, total: l.price });
  };

  const bookings: Booking[] = [
    lessonB(2, 1, 1, 0, 9, 'paid'),
    lessonB(6, 3, 1, 0, 11, 'paid'),
    lessonB(7, 1, 1, 0, 14, 'pending'),
    lessonB(2, 1, 1, 6, 10, 'paid'),
    lessonB(2, 2, 2, 8, 12, 'pending'),
    lessonB(2, 3, 1, -3, 9, 'completed'),
    lessonB(2, 4, 3, -5, 15, 'cancelled'),
    lessonB(7, 2, 2, -3, 9, 'completed'),
    lessonB(8, 4, 3, -5, 15, 'cancelled'),
    lessonB(6, 1, 1, 2, 16, 'paid'),
    lessonB(7, 5, 3, 3, 10, 'paid'),
    mk({ type: 'rental', status: 'pending', userId: 2, title: 'Аренда: комплект снаряжения', equipmentId: 3, size: '40', quantity: 1, start: at(addDays(today, 7), 14), end: at(addDays(today, 7), 16), total: 2800 }),
    mk({ type: 'rental', status: 'pending', userId: 6, title: 'Аренда: комплект снаряжения', equipmentId: 3, size: '42', quantity: 1, start: at(addDays(today, 7), 14), end: at(addDays(today, 7), 16), total: 2800 }),
  ];

  return { version: VERSION, seq, users, lessons, equipment, instructors, availability, bookings, feedback: [] };
}

let db: MockDb | null = null;

export function getDb(): MockDb {
  if (db) return db;
  try {
    const raw = localStorage.getItem(KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as MockDb;
      if (parsed.version === VERSION) {
        db = parsed;
        return db;
      }
    }
  } catch { /* ignore */ }
  db = seed();
  saveDb();
  return db;
}

export function saveDb() {
  try { localStorage.setItem(KEY, JSON.stringify(db)); } catch { /* ignore */ }
}

export function resetDb() {
  db = seed();
  saveDb();
}

export const nextId = () => ++getDb().seq;
