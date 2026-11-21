import type { BookingStatus, BookingType } from './types/index';

export function validRentalPeriod(start: string, end: string, now = Date.now()): boolean {
  const format = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/;
  if (!format.test(start) || !format.test(end)) return false;
  const a = Date.parse(start + ':00+03:00'), b = Date.parse(end + ':00+03:00');
  const hours = (b-a)/3600000;
  return Number.isFinite(a) && Number.isFinite(b) && a>now && hours>=1 && Number.isInteger(hours);
}

export function validQuantity(quantity: number, available: number): boolean {
  return Number.isInteger(quantity) && quantity>=1 && quantity<=available;
}

export function canCancelBooking(type: BookingType, status: BookingStatus, startsAt: string, now = Date.now()): boolean {
  return (status==='confirmed' || (type==='rental' && status==='pending')) && Date.parse(startsAt)>now;
}
