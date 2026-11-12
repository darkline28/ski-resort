import type { User } from '@/types';

/** Куда вести пользователя после входа / по кнопке «Личный кабинет» */
export const cabinetPath = (u: User | null) =>
  !u ? '/login' : u.role === 'admin' ? '/admin' : u.role === 'instructor' ? '/instructor' : '/account';
