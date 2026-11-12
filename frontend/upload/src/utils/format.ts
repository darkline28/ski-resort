import type { BookingStatus, EquipmentCategory, LessonCategory, Role } from '@/types';

export const fmtPrice = (n: number) => `${n.toLocaleString('ru-RU')} руб.`;

export const STATUS_LABEL: Record<BookingStatus, string> = {
  pending: 'Ожидание оплаты',
  paid: 'Оплачено',
  cancelled: 'Отменено',
  completed: 'Завершено',
};

export const ROLE_LABEL: Record<Role, string> = {
  client: 'Клиент',
  instructor: 'Инструктор',
  admin: 'Администратор',
};

export const LESSON_CATEGORY_LABEL: Record<LessonCategory, string> = {
  ski: 'Горные лыжи',
  snowboard: 'Сноуборд',
  individual: 'Индивидуальное обучение',
  group: 'Групповое занятие',
  kids: 'Детские занятия',
  freestyle: 'Фристайл',
};

export const EQUIPMENT_CATEGORY_LABEL: Record<EquipmentCategory, string> = {
  ski: 'Лыжи',
  snowboard: 'Сноуборды',
  boots: 'Ботинки',
  helmet: 'Шлемы',
  kit: 'Комплекты',
  goggles: 'Маски',
};

export const initials = (name: string) =>
  name.split(/\s+/).filter(Boolean).slice(0, 2).map((p) => p[0]!.toUpperCase()).join('');

/** «Алексей Кузнецов» → «Алексей К.» */
export const shortName = (name: string) => {
  const [first, last] = name.split(/\s+/);
  return last ? `${first} ${last[0]}.` : first;
};
