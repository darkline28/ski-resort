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

type BackendEquipment = Partial<Equipment> & {
  name?: string;
  price_per_hour?: number | string;
  image_url?: string | null;
  quantity?: number;
};

function equipmentImage(value?: string | null): string | null {
  if (!value) return null;
  if (/^(https?:\/\/|data:|\/images\/)/i.test(value)) return value;
  if (value.startsWith('images/')) return `/${value}`;
  if (value.startsWith('/')) return value;
  return `/images/${value}`;
}

function mapEquipment(item: BackendEquipment): Equipment {
  return {
    ...item,
    title: item.title ?? item.name ?? '',
    pricePerHour: Number(item.pricePerHour ?? item.price_per_hour ?? 0),
    stock: item.stock ?? item.quantity ?? 0,
    imageUrl: equipmentImage(item.imageUrl ?? item.image_url),
  } as Equipment;
}

type BackendInstructor = Partial<Instructor> & {
  experience_years?: number;
  is_active?: boolean;
  image_url?: string | null;
  description?: string | null;
};

function mapInstructor(item: BackendInstructor): Instructor {
  return {
    ...item,
    experienceYears: item.experienceYears ?? item.experience_years ?? 0,
    isActive: item.isActive ?? item.is_active ?? false,
    bio: item.bio ?? item.description ?? '',
    photoUrl: item.photoUrl ?? item.image_url ?? null,
  } as Instructor;
}

type BackendService = {
  id: number;
  name: string;
  description?: string | null;
  duration_minutes: number;
  price: number;
  is_active: boolean;
};

function mapService(item: BackendService): Lesson {
  return {
    id: item.id,
    title: item.name,
    description: item.description ?? '',
    durationMin: item.duration_minutes,
    price: item.price,
    active: item.is_active,
    category: 'individual',
    imageUrl: null,
  };
}

export const catalogApi = {
  lessons: async (_category?: LessonCategory) =>
  (await request<BackendService[]>('/services')).map(mapService),
  lesson: async (id: number) =>
  mapService(await request<BackendService>(`/services/${id}`)),
  equipment: async (category?: EquipmentCategory) => (await request<BackendEquipment[]>('/equipment', { query: { category } })).map(mapEquipment),
  equipmentItem: async (id: number) => mapEquipment(await request<BackendEquipment>(`/equipment/${id}`)),
  instructors: async () =>
  (await request<BackendInstructor[]>('/instructors')).map(mapInstructor),
  instructor: async (id: number) =>
  mapInstructor(await request<BackendInstructor>(`/instructors/${id}`)),
  slots: async (instructorId: number, date: string, lessonId: number) => {
  const data = await request<
    { starts_at: string; ends_at: string }[]
  >(`/instructors/${instructorId}/available-slots`, {
    query: {
      date,
      service_id: lessonId,
    },
  });

  return data.map((slot): Slot => ({
    start: slot.starts_at,
    end: slot.ends_at,
  }));
},
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

type BackendInstructorBooking = {
  id: number;
  client_name: string;
  service_name: string;
  starts_at: string;
  ends_at: string;
  status: Booking['status'];
};

function mapInstructorBooking(item: BackendInstructorBooking): Booking {
  return {
    id: item.id,
    clientName: item.client_name,
    title: item.service_name,
    start: item.starts_at,
    end: item.ends_at,
    status: item.status,
  } as Booking;
}

export const instructorApi = {
  me: () => request<Instructor>('/instructor/me'),
  bookings: async (date?: string) =>
  (
    await request<BackendInstructorBooking[]>(
      '/instructors/me/bookings',
      { query: { date } }
    )
  ).map(mapInstructorBooking),
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
