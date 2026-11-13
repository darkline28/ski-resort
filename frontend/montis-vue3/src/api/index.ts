/**
 * Все эндпоинты бэкенда в одном месте. Страницы импортируют только отсюда.
 * Согласуйте контракт с backend-разработчиком: docs/API.md.
 */
import { request } from './client';
import type {
  AdminStats, AuthResponse, Availability, AvailabilityPayload, Booking, BookingStatus, BookingType,
  CreateInstructorPayload, Equipment, EquipmentCategory, EquipmentInput, FeedbackPayload, Instructor,
  Lesson, LessonBookingPayload, LessonCategory, LessonInput, LoginPayload, PasswordPayload,
  PaymentPayload, ProfilePayload, RegisterPayload, RentalBookingPayload, Role, ScheduleCell, Slot,
  UpdateUserPayload, User,
} from '@/types';

export const authApi = {
  login: (p: LoginPayload) => request<AuthResponse>('/auth/login', { method: 'POST', body: p }),
  register: (p: RegisterPayload) => request<AuthResponse>('/auth/register', { method: 'POST', body: p }),
  logout: () => request<void>('/auth/logout', { method: 'POST' }),
  me: () => request<User>('/auth/me'),
};

export const catalogApi = {
  lessons: (category?: LessonCategory) => request<Lesson[]>('/lessons', { query: { category } }),
  lesson: (id: number) => request<Lesson>(`/lessons/${id}`),
  equipment: (category?: EquipmentCategory) => request<Equipment[]>('/equipment', { query: { category } }),
  equipmentItem: (id: number) => request<Equipment>(`/equipment/${id}`),
  instructors: (lessonId?: number) => request<Instructor[]>('/instructors', { query: { lessonId } }),
  instructor: (id: number) => request<Instructor>(`/instructors/${id}`),
  /** Свободные интервалы считает бэкенд */
  slots: (instructorId: number, date: string, lessonId: number) =>
    request<Slot[]>(`/instructors/${instructorId}/slots`, { query: { date, lessonId } }),
  /** Проверка наличия снаряжения на период (бэкенд) */
  equipmentAvailability: (id: number, start: string, end: string, size: string) =>
    request<{ available: number }>(`/equipment/${id}/availability`, { query: { start, end, size } }),
};

export const bookingApi = {
  createLesson: (p: LessonBookingPayload) => request<Booking>('/bookings/lesson', { method: 'POST', body: p }),
  createRental: (p: RentalBookingPayload) => request<Booking>('/bookings/rental', { method: 'POST', body: p }),
  get: (id: number) => request<Booking>(`/bookings/${id}`),
  my: (type?: BookingType) => request<Booking[]>('/me/bookings', { query: { type } }),
  cancel: (id: number) => request<Booking>(`/bookings/${id}/cancel`, { method: 'POST' }),
  pay: (id: number, p: PaymentPayload) => request<Booking>(`/bookings/${id}/pay`, { method: 'POST', body: p }),
};

export const profileApi = {
  update: (p: ProfilePayload) => request<User>('/me/profile', { method: 'PATCH', body: p }),
  changePassword: (p: PasswordPayload) => request<void>('/me/password', { method: 'POST', body: p }),
};

export const feedbackApi = {
  send: (p: FeedbackPayload) => request<void>('/feedback', { method: 'POST', body: p }),
};

export const instructorApi = {
  me: () => request<Instructor>('/instructor/me'),
  bookings: (from: string, to: string) => request<Booking[]>('/instructor/bookings', { query: { from, to } }),
  schedule: (from: string, to: string) => request<ScheduleCell[]>('/instructor/schedule', { query: { from, to } }),
  addAvailability: (p: AvailabilityPayload) =>
    request<Availability>('/instructor/availability', { method: 'POST', body: p }),
  removeAvailability: (id: number) => request<void>(`/instructor/availability/${id}`, { method: 'DELETE' }),
  complete: (bookingId: number) => request<Booking>(`/bookings/${bookingId}/complete`, { method: 'POST' }),
};

export const adminApi = {
  stats: () => request<AdminStats>('/admin/stats'),
  users: (role?: Role) => request<User[]>('/admin/users', { query: { role } }),
  updateUser: (id: number, p: UpdateUserPayload) => request<User>(`/admin/users/${id}`, { method: 'PATCH', body: p }),
  createInstructor: (p: CreateInstructorPayload) =>
    request<Instructor>('/admin/instructors', { method: 'POST', body: p }),

  lessons: () => request<Lesson[]>('/admin/lessons'),
  createLesson: (p: LessonInput) => request<Lesson>('/admin/lessons', { method: 'POST', body: p }),
  updateLesson: (id: number, p: Partial<LessonInput>) =>
    request<Lesson>(`/admin/lessons/${id}`, { method: 'PATCH', body: p }),
  deleteLesson: (id: number) => request<void>(`/admin/lessons/${id}`, { method: 'DELETE' }),

  equipment: () => request<Equipment[]>('/admin/equipment'),
  createEquipment: (p: EquipmentInput) => request<Equipment>('/admin/equipment', { method: 'POST', body: p }),
  updateEquipment: (id: number, p: Partial<EquipmentInput>) =>
    request<Equipment>(`/admin/equipment/${id}`, { method: 'PATCH', body: p }),
  deleteEquipment: (id: number) => request<void>(`/admin/equipment/${id}`, { method: 'DELETE' }),

  bookings: (status?: BookingStatus) => request<Booking[]>('/admin/bookings', { query: { status } }),
  setBookingStatus: (id: number, status: BookingStatus) =>
    request<Booking>(`/admin/bookings/${id}`, { method: 'PATCH', body: { status } }),
};
