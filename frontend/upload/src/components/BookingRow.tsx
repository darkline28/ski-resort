import type { ReactNode } from 'react';
import type { Booking } from '@/types';
import { durationMinutes, fmtDayLong, fmtDuration, fmtTime } from '@/utils/date';
import { fmtPrice, shortName } from '@/utils/format';
import { StatusBadge } from './ui';

/** Строка брони в списках ЛК: время · название · дата · статус · действия */
export function BookingRow({ booking: b, meta = 'duration', actions }: {
  booking: Booking;
  meta?: 'duration' | 'price';
  actions?: ReactNode;
}) {
  const title = b.instructorName ? `${b.title}, ${shortName(b.instructorName)}` : b.title;
  const sub = meta === 'price' ? fmtPrice(b.total) : fmtDuration(durationMinutes(b.start, b.end));
  return (
    <div className="booking-row">
      <div className="booking-row__time">{fmtTime(b.start)}</div>
      <div className="booking-row__info">
        <div className="booking-row__title">{title}{b.quantity && b.quantity > 1 ? ` × ${b.quantity}` : ''}</div>
        <div className="booking-row__sub">{fmtDayLong(b.start)} · {sub}</div>
      </div>
      <div className="booking-row__side">
        <StatusBadge status={b.status} />
        {actions && <div className="booking-row__actions">{actions}</div>}
      </div>
    </div>
  );
}
