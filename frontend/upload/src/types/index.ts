// Доменные типы. Совпадают с контрактом API (см. docs/API.md).

export type Role = 'client' | 'instructor' | 'admin';

export interface User {
  id: number;
  name: string;
  phone: string;
  email: string;
  role: Role;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export type LessonCategory = 'ski' | 'snowboard' | 'individual' | 'group' | 'kids' | 'freestyle';

/** Занятие (услуга) */
export interface Lesson {
  id: number;
  title: string;
  description: string;
  category: LessonCategory;
  price: number; // руб. за занятие (минимальная цена — «от»)
  durationMin: number;
  imageUrl?: string | null;
  active: boolean;
}

export type EquipmentCategory = 'ski' | 'snowboard' | 'boots' | 'helmet' | 'kit' | 'goggles';

/** Позиция проката */
export interface Equipment {
  id: number;
  title: string;
  description: string;
  category: EquipmentCategory;
  pricePerHour: number;
  sizes: string[];
  stock: number; // доступное количество единиц
  imageUrl?: string | null;
  active: boolean;
}

export interface Instructor {
  id: number;
  userId: number;
  name: string;
  email: string;
  specialization: string;
  /** категории занятий, которые ведёт инструктор */
  categories: LessonCategory[];
  experienceYears: number;
  pricePerHour: number;
  bio: string;
  photoUrl?: string | null;
}

/** Свободный/занятый интервал, который рассчитывает бэкенд */
export interface Slot {
  start: string; // ISO
  end: string; // ISO
  available: boolean;
}

export type BookingStatus = 'pending' | 'paid' | 'cancelled' | 'completed';
export type BookingType = 'lesson' | 'rental';

export interface Booking {
  id: number;
  type: BookingType;
  status: BookingStatus;
  userId: number;
  clientName: string;
  title: string; // «Горные лыжи», «Аренда: комплект снаряжения»
  lessonId?: number;
  instructorId?: number;
  instructorName?: string;
  equipmentId?: number;
  size?: string;
  quantity?: number;
  start: string; // ISO
  end: string; // ISO
  total: number;
  createdAt: string;
}

/** Окно свободного времени инструктора */
export interface Availability {
  id: number;
  instructorId: number;
  start: string;
  end: string;
}

/** Ячейка недельного расписания инструктора */
export interface ScheduleCell {
  start: string;
  end: string;
  status: 'free' | 'booked';
  availabilityId?: number;
  bookingId?: number;
}

export interface AdminStats {
  users: number;
  instructors: number;
  bookingsToday: number;
  revenueWeek: number;
  bookingsByWeekday: number[]; // Пн..Вс
  latestBookings: Booking[];
}

// ---- payloads ----
export interface LoginPayload { email: string; password: string }
export interface RegisterPayload { name: string; phone: string; email: string; password: string }
export interface LessonBookingPayload { lessonId: number; instructorId: number; start: string }
export interface RentalBookingPayload { equipmentId: number; size: string; quantity: number; start: string; end: string }
export interface PaymentPayload { cardNumber: string; expiry: string; cvc: string }
export interface ProfilePayload { name: string; phone: string; email: string }
export interface PasswordPayload { currentPassword: string; newPassword: string }
export interface FeedbackPayload { name: string; contact: string; message: string }
export interface AvailabilityPayload { start: string; end: string }
export interface CreateInstructorPayload {
  name: string;
  email: string;
  password: string;
  specialization: string;
}
export interface UpdateUserPayload { name?: string; phone?: string; email?: string; role?: Role }
export type LessonInput = Omit<Lesson, 'id'>;
export type EquipmentInput = Omit<Equipment, 'id'>;
