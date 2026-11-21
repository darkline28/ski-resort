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

export type EquipmentCategory =
  | 'ski'
  | 'snowboard'
  | 'jacket'
  | 'pants'
  | 'boots'
  | 'helmet'
  | 'goggles'
  | 'kit' | (string & {});

export interface Equipment {
  id: number;
  title: string;
  description: string;
  category: EquipmentCategory;
  pricePerHour: number;
  stock: number;
  imageUrl?: string | null;
  active: boolean;
}

export interface Instructor {
  id: number;
  userId?: number;
  name: string;
  email?: string;
  specialization: string;
  categories?: LessonCategory[];
  experienceYears: number;
  pricePerHour?: number;
  bio?: string | null;
  photoUrl?: string | null;
  isActive: boolean;
}

export type BookingStatus =
  | 'pending'
  | 'confirmed'
  | 'cancelled'
  | 'completed'
  | 'no_show'
  | 'issued'
  | 'returned';

export type BookingType = 'lesson' | 'rental';

export interface Booking {
  id: number;
  type: BookingType;
  status: BookingStatus;
  userId?: number;
  clientName?: string;
  title: string;
  lessonId?: number;
  instructorId?: number;
  instructorName?: string;
  equipmentId?: number;
  size?: string;
  quantity?: number;
  start: string;
  end: string;
  total?: number;
  createdAt?: string;
}

export interface AuthResponse { access_token: string; token_type: 'Bearer'; user: User }
export interface LoginPayload { email: string; password: string }
export interface RegisterPayload extends LoginPayload { name: string; phone: string }
export interface Slot { start: string; end: string }
export interface LessonBookingPayload { lessonId: number; instructorId: number; start: string }
export interface RentalBookingPayload { equipmentId: number; size: string; quantity: number; start: string; end: string }
