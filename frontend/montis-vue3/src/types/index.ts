export type Role = 'client' | 'instructor' | 'admin';

export interface User {
  id: number;
  name: string;
  phone: string;
  email: string;
  role: Role;
}

export type LessonCategory = 'ski' | 'snowboard' | 'individual' | 'group' | 'kids' | 'freestyle';

export interface Lesson {
  id: number;
  title: string;
  description: string;
  category: LessonCategory;
  price: number;
  durationMin: number;
  imageUrl?: string | null;
  active: boolean;
}

export type EquipmentCategory = 'ski' | 'snowboard' | 'boots' | 'helmet' | 'kit' | 'goggles';

export interface Equipment {
  id: number;
  title: string;
  description: string;
  category: EquipmentCategory;
  pricePerHour: number;
  sizes: string[];
  stock: number;
  imageUrl?: string | null;
  active: boolean;
}

export interface Instructor {
  id: number;
  userId: number;
  name: string;
  email: string;
  specialization: string;
  categories: LessonCategory[];
  experienceYears: number;
  pricePerHour: number;
  bio: string;
  photoUrl?: string | null;
}

export type BookingStatus = 'pending' | 'paid' | 'cancelled' | 'completed';
export type BookingType = 'lesson' | 'rental';

export interface Booking {
  id: number;
  type: BookingType;
  status: BookingStatus;
  userId: number;
  clientName: string;
  title: string;
  lessonId?: number;
  instructorId?: number;
  instructorName?: string;
  equipmentId?: number;
  size?: string;
  quantity?: number;
  start: string;
  end: string;
  total: number;
  createdAt: string;
}

// Предварительные DTO для согласования с backend-разработчиком.
export interface AuthResponse { token: string; user: User }
export interface LoginPayload { email: string; password: string }
export interface RegisterPayload extends LoginPayload { name: string; phone: string }
export interface Slot { start: string; end: string }
export interface LessonBookingPayload { lessonId: number; instructorId: number; start: string }
export interface RentalBookingPayload { equipmentId: number; size: string; quantity: number; start: string; end: string }
export interface PaymentPayload { /* платёжный провайдер определяется командой */ }
export interface ProfilePayload { name: string; phone: string; email: string }
export interface PasswordPayload { current: string; next: string }
export interface FeedbackPayload { name: string; email: string; message: string }
export interface LessonInput extends Omit<Lesson, 'id'> {}
export interface EquipmentInput extends Omit<Equipment, 'id'> {}
export interface UpdateUserPayload { name?: string; phone?: string; email?: string; role?: Role }
export interface CreateInstructorPayload { name: string; email: string; password: string; specialization: string }
export interface Availability { id: number; start: string; end: string }
export interface AvailabilityPayload { start: string; end: string }
export interface ScheduleCell { start: string; end: string; available: boolean; bookingId?: number }
export interface AdminStats { users: number; bookings: number; lessons: number; equipment: number }
