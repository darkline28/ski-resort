// Работа с датами без сторонних библиотек. Время показываем в локальной зоне браузера.

const pad = (n: number) => String(n).padStart(2, '0');

/** YYYY-MM-DD в локальной зоне */
export function toDateKey(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

/** Локальная дата+время → ISO-строка (с учётом смещения) */
export function localIso(dateKey: string, time: string): string {
  const [y, m, d] = dateKey.split('-').map(Number);
  const [hh, mm] = time.split(':').map(Number);
  return new Date(y, m - 1, d, hh, mm || 0).toISOString();
}

export function addDays(d: Date, n: number): Date {
  const r = new Date(d);
  r.setDate(r.getDate() + n);
  return r;
}

export function startOfDay(d: Date): Date {
  const r = new Date(d);
  r.setHours(0, 0, 0, 0);
  return r;
}

/** Понедельник недели, в которую попадает дата */
export function startOfWeek(d: Date): Date {
  const r = startOfDay(d);
  const day = (r.getDay() + 6) % 7; // Пн = 0
  return addDays(r, -day);
}

export function isSameDay(a: Date, b: Date): boolean {
  return toDateKey(a) === toDateKey(b);
}

export const fmtTime = (iso: string) =>
  new Date(iso).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });

/** «Ср, 18 ноября» */
export const fmtDayLong = (iso: string | Date) => {
  const d = new Date(iso);
  const wd = d.toLocaleDateString('ru-RU', { weekday: 'short' });
  const rest = d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
  return `${cap(wd)}, ${rest}`;
};

/** «18 ноября» */
export const fmtDay = (iso: string | Date) =>
  new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });

/** «18.11 10:00» */
export const fmtShort = (iso: string) => {
  const d = new Date(iso);
  return `${pad(d.getDate())}.${pad(d.getMonth() + 1)} ${fmtTime(iso)}`;
};

/** «Пн» */
export const fmtWeekdayShort = (d: Date) => cap(d.toLocaleDateString('ru-RU', { weekday: 'short' }));

export const cap = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);

export function durationMinutes(start: string, end: string): number {
  return Math.round((new Date(end).getTime() - new Date(start).getTime()) / 60000);
}

/** «60 мин», «2 часа», «1 ч 30 мин» */
export function fmtDuration(min: number): string {
  if (min <= 90) return `${min} мин`;
  if (min % 60 === 0) {
    const h = min / 60;
    return `${h} ${plural(h, ['час', 'часа', 'часов'])}`;
  }
  return `${Math.floor(min / 60)} ч ${min % 60} мин`;
}

export function plural(n: number, forms: [string, string, string]): string {
  const n10 = n % 10, n100 = n % 100;
  if (n10 === 1 && n100 !== 11) return forms[0];
  if (n10 >= 2 && n10 <= 4 && (n100 < 10 || n100 >= 20)) return forms[1];
  return forms[2];
}

/** Значение для <input type="datetime-local"> */
export function toInputDateTime(d: Date): string {
  return `${toDateKey(d)}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
