/**
 * HTTP-клиент. Все запросы к бэкенду идут через request().
 * Если VITE_USE_MOCKS=true — запрос обрабатывает локальная заглушка (src/api/mock),
 * которая повторяет контракт docs/API.md. Переключение не требует правок в страницах.
 */
// Заглушка подгружается лениво и не попадает в бандл, если VITE_USE_MOCKS=false

export const API_URL = (import.meta.env.VITE_API_URL || '/api').replace(/\/$/, '');
export const USE_MOCKS = import.meta.env.VITE_USE_MOCKS !== 'false';

const TOKEN_KEY = 'montis_token';

export const tokenStorage = {
  get: (): string | null => {
    try { return localStorage.getItem(TOKEN_KEY); } catch { return null; }
  },
  set: (token: string) => {
    try { localStorage.setItem(TOKEN_KEY, token); } catch { /* ignore */ }
  },
  clear: () => {
    try { localStorage.removeItem(TOKEN_KEY); } catch { /* ignore */ }
  },
};

export class ApiError extends Error {
  status: number;
  /** ошибки валидации по полям: { email: 'Уже занят' } */
  fields?: Record<string, string>;

  constructor(status: number, message: string, fields?: Record<string, string>) {
    super(message);
    this.status = status;
    this.fields = fields;
  }
}

type Method = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';

export interface RequestOptions {
  method?: Method;
  body?: unknown;
  query?: Record<string, string | number | undefined | null>;
  signal?: AbortSignal;
}

/** Обработчик 401 — выставляется AuthContext, чтобы разлогинить пользователя */
let onUnauthorized: (() => void) | null = null;
export const setUnauthorizedHandler = (fn: (() => void) | null) => { onUnauthorized = fn; };

function buildQuery(query?: RequestOptions['query']): string {
  if (!query) return '';
  const params = new URLSearchParams();
  Object.entries(query).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') params.append(k, String(v));
  });
  const s = params.toString();
  return s ? `?${s}` : '';
}

export async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const method = opts.method ?? 'GET';
  const url = path + buildQuery(opts.query);
  const token = tokenStorage.get();

  let status: number;
  let data: unknown;

  if (USE_MOCKS) {
    const { mockFetch } = await import('./mock/server');
    const res = await mockFetch(method, url, opts.body, token);
    status = res.status;
    data = res.body;
  } else {
    let res: Response;
    try {
      res = await fetch(API_URL + url, {
        method,
        signal: opts.signal,
        headers: {
          Accept: 'application/json',
          ...(opts.body !== undefined ? { 'Content-Type': 'application/json' } : {}),
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
      });
    } catch (e) {
      if ((e as Error).name === 'AbortError') throw e;
      throw new ApiError(0, 'Нет соединения с сервером. Проверь интернет и попробуй ещё раз.');
    }
    status = res.status;
    const text = await res.text();
    try { data = text ? JSON.parse(text) : null; } catch { data = text; }
  }

  if (status >= 200 && status < 300) return data as T;

  const err = (data ?? {}) as { message?: string; errors?: Record<string, string> };
  if (status === 401 && onUnauthorized) onUnauthorized();
  throw new ApiError(status, err.message || defaultMessage(status), err.errors);
}

function defaultMessage(status: number): string {
  switch (status) {
    case 400: return 'Проверь правильность заполнения полей.';
    case 401: return 'Нужно войти в аккаунт.';
    case 403: return 'Недостаточно прав для этого действия.';
    case 404: return 'Данные не найдены.';
    case 409: return 'Конфликт данных: интервал уже занят.';
    case 422: return 'Проверь правильность заполнения полей.';
    default: return 'Не удалось загрузить данные. Попробуй ещё раз.';
  }
}
