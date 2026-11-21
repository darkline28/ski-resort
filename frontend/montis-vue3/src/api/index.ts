import { request } from './client';
import type {
  AuthResponse, Booking, BookingStatus, BookingType, Equipment, EquipmentCategory, Instructor,
  Lesson, LessonBookingPayload, LessonCategory, LoginPayload, RegisterPayload, RentalBookingPayload, Slot, User,
} from '@/types';

export const authApi = {
  login: (p: LoginPayload) => request<AuthResponse>('/auth/login', { method: 'POST', body: p }),
  register: (p: RegisterPayload) => request<AuthResponse>('/auth/register', { method: 'POST', body: p }),
  me: () => request<User>('/auth/me'),
};

type BackendEquipment = {
  id:number;
  name: string;
  category:string;
  description?:string;
  price_per_hour: number;
  image_url?: string | null;
  quantity: number;
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
    id:item.id,
    category:item.category,
    title: item.name,
    pricePerHour: item.price_per_hour,
    stock: item.quantity,
    imageUrl: equipmentImage(item.image_url),
    active: true,
    description: item.description ?? '',
  };
}

type BackendInstructor = {
  id:number;
  name:string;
  specialization:string;
  experience_years: number;
  is_active: boolean;
  image_url?: string | null;
  description?: string | null;
};

function mapInstructor(item: BackendInstructor): Instructor {
  return {
    id:item.id,
    name:item.name,
    specialization:item.specialization,
    experienceYears: item.experience_years,
    isActive: item.is_active,
    bio: item.description ?? '',
    photoUrl: equipmentImage(item.image_url),
  };
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
  
equipment: async (category?: EquipmentCategory) =>
  (
    await request<BackendEquipment[]>('/equipment', {
      query: { category },
    })
  ).map(mapEquipment),

equipmentItem: async (id: number) =>
  mapEquipment(
    await request<BackendEquipment>(`/equipment/${id}`)
  ),

equipmentCategories: () =>
  request<{ id: number; name: string; code: string }[]>(
    '/equipment/categories'
  ),

equipmentSets: () =>
  request<{
    id: number;
    name: string;
    description?: string;
    sport_type: 'ski' | 'snowboard';
    age_group: 'adult' | 'child';
    price_per_hour: number;
    is_active: boolean;
  }[]>('/equipment/sets'),

instructors: async () =>

  (await request<BackendInstructor[]>('/instructors')).map(mapInstructor),
  instructor: async (id: number) =>
  mapInstructor(await request<BackendInstructor>(`/instructors/${id}`)),
  instructorServices: async (instructorId: number) =>
  (
    await request<BackendService[]>(
      `/instructors/${instructorId}/services`
    )
  ).map(mapService),
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
  equipmentSet: (id: number) => request<{id:number;name:string;description?:string;price_per_hour:number;is_active:boolean;items:{category_id:number;category_name:string;quantity:number}[]}>(`/equipment/sets/${id}`),
  equipmentAvailability: (id: number | undefined, start: string, end: string) =>
    request<{ equipment_id: number; size: string; available_quantity: number }[]>('/equipment/availability', { query: { equipment_id: id, starts_at: start, ends_at: end } }),
};

type BackendLessonBooking = {
  id: number;
  user_id: number;
  instructor_id: number;
  service_id: number;
  created_at: string;
  instructor_name: string;
  service_name: string;
  starts_at: string;
  ends_at: string;
  status: Booking['status'];
  total_price: number;
};

type BackendRentalBooking = {
  id: number;
  user_id: number;
  created_at: string;
  starts_at: string;
  ends_at: string;
  status: 'pending' | 'confirmed' | 'issued' | 'returned' | 'cancelled';
  total_price: number;
  items: { name: string; size: string; booked_set_id: number | null }[];
  sets: { name: string }[];
};

type AccountBooking = Booking & {
  bookingType: 'lesson' | 'rental';
};

function mapLessonBooking(b: BackendLessonBooking): AccountBooking {
  return {
    id: b.id,
    type: 'lesson',
    bookingType: 'lesson',
    title: `${b.service_name} — ${b.instructor_name}`,
    start: b.starts_at,
    end: b.ends_at,
    status: b.status,
    total: b.total_price,
    userId: b.user_id,
    createdAt: b.created_at,
    instructorId: b.instructor_id,
    lessonId: b.service_id,
  };
}

function mapRentalBooking(b: BackendRentalBooking): AccountBooking {
  const names = [
    ...(b.items ?? []).filter(item=>item.booked_set_id===null).map((item) => `${item.name} (${item.size})`),
    ...(b.sets ?? []).map((set) => set.name),
  ];

  return {
    id: b.id,
    type: 'rental',
    bookingType: 'rental',
    title: names.join(', ') || 'Аренда оборудования',
    start: b.starts_at,
    end: b.ends_at,
    status: b.status,
    total: b.total_price,
    userId: b.user_id,
    createdAt: b.created_at,
  };
}

export const bookingApi = {
  createSet: (id: number, quantity: number, sizes: {category_id:number;size:string}[], start: string, end: string) =>
    request<unknown>('/equipment-bookings', {method:'POST', body:{starts_at:start,ends_at:end,sets:[{equipment_set_id:id,quantity,sizes}]}}),
  createLesson: (p: LessonBookingPayload) =>
  request<Omit<BackendLessonBooking,'instructor_name'|'service_name'>>('/lesson-bookings', {
    method: 'POST',
    body: {
      instructor_id: p.instructorId,
      service_id: p.lessonId,
      starts_at: p.start,
    },
  }),
  
createRental: (p: RentalBookingPayload) =>
  request<BackendRentalBooking>('/equipment-bookings', {
    method: 'POST',
    body: {
      starts_at: p.start,
      ends_at: p.end,
      items: [
        {
          equipment_id: p.equipmentId,
          size: p.size,
          quantity: p.quantity,
        },
      ],
      sets: [],
    },
  }),

  get: async (id: number, type: BookingType) => type === 'lesson'
    ? mapLessonBooking(await request<BackendLessonBooking>(`/lesson-bookings/${id}`))
    : mapRentalBooking(await request<BackendRentalBooking>(`/equipment-bookings/${id}`)),
  my: async (): Promise<{ bookings: AccountBooking[]; errors: string[] }> => {
    const [lessons, rentals] = await Promise.allSettled([
      request<BackendLessonBooking[]>('/lesson-bookings/my'),
      request<BackendRentalBooking[]>('/equipment-bookings/my'),
    ]);
    const bookings = [
      ...(lessons.status === 'fulfilled' ? lessons.value.map(mapLessonBooking) : []),
      ...(rentals.status === 'fulfilled' ? rentals.value.map(mapRentalBooking) : []),
    ].sort((a, b) => new Date(b.start).getTime() - new Date(a.start).getTime());
    const errors = [
      ...(lessons.status === 'rejected' ? ['Не удалось загрузить занятия'] : []),
      ...(rentals.status === 'rejected' ? ['Не удалось загрузить аренду'] : []),
    ];
    return { bookings, errors };
  },
  cancel: (id: number, type: 'lesson' | 'rental') =>
  request<BackendLessonBooking | BackendRentalBooking>(
    type === 'lesson'
      ? `/lesson-bookings/${id}/cancel`
      : `/equipment-bookings/${id}/cancel`,
    { method: 'PATCH' }
  ),
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
    type:'lesson',
    clientName: item.client_name,
    title: item.service_name,
    start: item.starts_at,
    end: item.ends_at,
    status: item.status,
  };
}

export const instructorApi = {
  bookings: async (date?: string) =>
  (
    await request<BackendInstructorBooking[]>(
      '/instructors/me/bookings',
      { query: { date } }
    )
  ).map(mapInstructorBooking),
  setStatus: async (id: number, status: 'completed' | 'no_show') =>
    mapInstructorBooking(await request<BackendInstructorBooking>(`/instructors/me/bookings/${id}/status`, { method: 'PATCH', body: { status } })),
};

export const adminApi = {
  clients: () => request<{id:number;name:string;email:string;phone:string}[]>('/admin/clients'),
  instructors: async () => (await request<BackendInstructor[]>('/admin/instructors')).map(mapInstructor),
  services: async () => (await request<BackendService[]>('/admin/services')).map(mapService),
  instructorServices: async (id:number) => (await request<BackendService[]>(`/admin/instructors/${id}/services`)).map(mapService),
  assignService: (id:number, service_id:number) => request<void>(`/admin/instructors/${id}/services`,{method:'POST',body:{service_id}}),
  removeService: (id:number, serviceId:number) => request<void>(`/admin/instructors/${id}/services/${serviceId}`,{method:'DELETE'}),
  appointInstructor: (body:{user_id:number;specialization:string;experience_years:number;description?:string;image_url?:string}) => request<BackendInstructor>('/admin/instructors',{method:'POST',body}),
  units: (equipment_id?: number) => request<Record<string,unknown>[]>('/admin/equipment/units',{query:{equipment_id}}),
  schedules: (id: number) => request<Record<string,unknown>[]>(`/admin/instructors/${id}/schedules`),
  save: (path: string, body: Record<string,unknown>, id?: number) => request<unknown>(`/admin/${path}${id ? '/'+id : ''}`,{method:id?'PATCH':'POST',body}),
  remove: (path: string, id: number) => request<void>(`/admin/${path}/${id}`,{method:'DELETE'}),
  bookings: async (type: BookingType) => type === 'lesson'
    ? (await request<(BackendLessonBooking & {client_name: string})[]>('/admin/lesson-bookings')).map(b => ({...mapLessonBooking(b), clientName: b.client_name}))
    : (await request<(BackendRentalBooking & {client_name: string})[]>('/admin/equipment-bookings')).map(b => ({...mapRentalBooking(b), clientName: b.client_name})),
  setBookingStatus: (id: number, type: BookingType, status: BookingStatus) =>
    request<unknown>(`/admin/${type === 'lesson' ? 'lesson' : 'equipment'}-bookings/${id}`, { method: 'PATCH', body: { status } }),
};
