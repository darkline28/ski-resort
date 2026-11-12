import { useMemo, useState } from 'react';
import { instructorApi } from '@/api';
import { ApiError } from '@/api/client';
import { useAsync } from '@/hooks/useAsync';
import { useToast } from '@/context/ToastContext';
import { PageHead } from '@/layouts/CabinetLayout';
import { AsyncView, EmptyState } from '@/components/States';
import { Alert, Button, Modal, Photo, StatusBadge } from '@/components/ui';
import type { Booking, ScheduleCell } from '@/types';
import { addDays, durationMinutes, fmtDay, fmtDuration, fmtTime, fmtWeekdayShort, isSameDay, localIso, startOfDay, startOfWeek, toDateKey } from '@/utils/date';
import { ProfileForm, PasswordForm } from '../account/ClientPages';

const HOURS = [9, 10, 11, 12, 13, 14, 15, 16, 17];

// =================== Экран 8: главная инструктора ===================
export function InstructorHomePage() {
  const today = startOfDay(new Date());
  const [day, setDay] = useState(0);
  const date = addDays(today, day);
  const bookings = useAsync(
    () => instructorApi.bookings(date.toISOString(), addDays(date, 1).toISOString()),
    [day],
  );

  return (
    <>
      <PageHead title="Твоё расписание на сегодня" text="Здесь отображаются ближайшие занятия и актуальная информация о записях клиентов." />
      <div className="card panel">
        <div className="panel__head">
          <h3 className="h3">{day === 0 ? 'Сегодня' : day === 1 ? 'Завтра' : fmtWeekdayShort(date)}, {fmtDay(date)}</h3>
          <div className="pager">
            <button className="pager__btn" onClick={() => setDay((d) => d - 1)} aria-label="Предыдущий день">‹</button>
            <button className="pager__btn" onClick={() => setDay(0)} disabled={day === 0}>Сегодня</button>
            <button className="pager__btn" onClick={() => setDay((d) => d + 1)} aria-label="Следующий день">›</button>
          </div>
        </div>
        <AsyncView state={bookings} empty={<EmptyState title="Занятий нет" text="На этот день записей клиентов пока нет." />}>
          {(list) => (
            <div className="list">
              {list.map((b) => (
                <div className="booking-row" key={b.id}>
                  <div className="booking-row__time">{fmtTime(b.start)}</div>
                  <div className="booking-row__info">
                    <div className="booking-row__title">{b.clientName} · {b.title}</div>
                    <div className="booking-row__sub">{fmtDuration(durationMinutes(b.start, b.end))}</div>
                  </div>
                  <div className="booking-row__side"><StatusBadge status={b.status} /></div>
                </div>
              ))}
            </div>
          )}
        </AsyncView>
      </div>
    </>
  );
}

// =================== Экран 9: расписание и занятия ===================
export function InstructorSchedulePage() {
  const toast = useToast();
  const [weekOffset, setWeekOffset] = useState(0);
  const weekStart = useMemo(() => addDays(startOfWeek(new Date()), weekOffset * 7), [weekOffset]);
  const weekEnd = addDays(weekStart, 7);
  const days = Array.from({ length: 7 }, (_, i) => addDays(weekStart, i));

  const schedule = useAsync(() => instructorApi.schedule(weekStart.toISOString(), weekEnd.toISOString()), [weekOffset]);
  const bookings = useAsync(() => instructorApi.bookings(weekStart.toISOString(), weekEnd.toISOString()), [weekOffset]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [modal, setModal] = useState(false);
  const [busyCell, setBusyCell] = useState<string | null>(null);

  const cellAt = (d: Date, h: number): ScheduleCell | undefined =>
    schedule.data?.find((c) => { const s = new Date(c.start); return isSameDay(s, d) && s.getHours() === h; });

  /** Клик по ячейке: пусто → добавить свободный час, свободно → убрать, занято → показать запись */
  const toggle = async (d: Date, h: number) => {
    const cell = cellAt(d, h);
    const start = new Date(d); start.setHours(h, 0, 0, 0);
    if (start.getTime() < Date.now()) return;
    if (cell?.status === 'booked') { setSelectedId(cell.bookingId ?? null); return; }
    const key = start.toISOString();
    setBusyCell(key);
    try {
      if (cell?.availabilityId) await instructorApi.removeAvailability(cell.availabilityId);
      else await instructorApi.addAvailability({ start: key, end: new Date(start.getTime() + 3600000).toISOString() });
      schedule.reload();
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Не удалось обновить расписание', 'error');
    } finally {
      setBusyCell(null);
    }
  };

  const complete = async (b: Booking) => {
    try {
      const updated = await instructorApi.complete(b.id);
      bookings.setData((list) => (list ?? []).map((x) => (x.id === b.id ? updated : x)));
      toast('Занятие отмечено завершённым');
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Ошибка', 'error');
    }
  };

  const weekLabel = `${weekStart.getDate()}–${fmtDay(addDays(weekStart, 6))}`;
  const selected = bookings.data?.find((b) => b.id === selectedId) ?? null;

  return (
    <>
      <PageHead title="Управление занятиями" text="Просматривай записи клиентов, управляй расписанием и отмечай свободное время для новых занятий." />
      <div className="schedule-layout">
        <div className="card panel">
          <div className="panel__head">
            <div className="pager">
              <button className="pager__btn" onClick={() => setWeekOffset((w) => w - 1)} aria-label="Предыдущая неделя">‹</button>
              <h3 className="h3">Неделя {weekLabel}</h3>
              <button className="pager__btn" onClick={() => setWeekOffset((w) => w + 1)} aria-label="Следующая неделя">›</button>
            </div>
            <Button onClick={() => setModal(true)}>Добавить свободное время</Button>
          </div>
          <AsyncView state={schedule}>
            {() => (
              <div className="scroll-x">
                <table className="week">
                  <thead>
                    <tr><th />{days.map((d) => <th key={d.toISOString()} className={isSameDay(d, new Date()) ? 'week__today' : ''}>{fmtWeekdayShort(d)} {d.getDate()}</th>)}</tr>
                  </thead>
                  <tbody>
                    {HOURS.map((h) => (
                      <tr key={h}>
                        <th className="week__hour">{String(h).padStart(2, '0')}:00</th>
                        {days.map((d) => {
                          const c = cellAt(d, h);
                          const start = new Date(d); start.setHours(h);
                          const past = start.getTime() < Date.now();
                          const busy = busyCell === new Date(new Date(d).setHours(h, 0, 0, 0)).toISOString();
                          return (
                            <td key={d.toISOString()}>
                              <button
                                className={`slot slot--${c?.status ?? 'none'} ${past ? 'slot--past' : ''} ${c?.bookingId && c.bookingId === selectedId ? 'slot--selected' : ''}`}
                                onClick={() => toggle(d, h)}
                                disabled={busy || (past && c?.status !== 'booked')}
                                title={c?.status === 'booked' ? 'Запись клиента' : c ? 'Нажми, чтобы убрать' : 'Нажми, чтобы открыть для записи'}
                              >
                                {busy ? '…' : c?.status === 'booked' ? 'Занято' : c ? 'Свободно' : ''}
                              </button>
                            </td>
                          );
                        })}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </AsyncView>
          <p className="muted small">Нажми на пустую ячейку, чтобы открыть время для записи, или на «Свободно», чтобы закрыть его.</p>
        </div>

        <div className="card panel">
          <h3 className="h3">Записи клиентов</h3>
          <AsyncView state={bookings} empty={<EmptyState title="Записей нет" text="На этой неделе пока никто не записался." />}>
            {(list) => (
              <div className="list">
                {list.map((b) => (
                  <button key={b.id} className={`client-row ${selectedId === b.id ? 'client-row--active' : ''}`} onClick={() => setSelectedId(b.id)}>
                    <div className="client-row__when">{fmtWeekdayShort(new Date(b.start))}<br />{fmtTime(b.start)}</div>
                    <div className="client-row__info">
                      <div className="booking-row__title">{b.clientName}</div>
                      <div className="booking-row__sub">{b.title}</div>
                    </div>
                    <StatusBadge status={b.status} />
                  </button>
                ))}
              </div>
            )}
          </AsyncView>
          {selected && (
            <div className="selected-booking">
              <div className="muted small">{fmtDay(selected.start)}, {fmtTime(selected.start)}–{fmtTime(selected.end)}</div>
              <Button variant="light" block disabled={selected.status !== 'paid' || new Date(selected.start) > new Date()}
                onClick={() => complete(selected)}>
                Отметить занятие завершённым
              </Button>
              {selected.status === 'paid' && new Date(selected.start) > new Date() && <p className="muted small">Можно завершить после начала занятия.</p>}
            </div>
          )}
        </div>
      </div>

      <AvailabilityModal open={modal} onClose={() => setModal(false)} onSaved={() => { setModal(false); schedule.reload(); }} />
    </>
  );
}

function AvailabilityModal({ open, onClose, onSaved }: { open: boolean; onClose: () => void; onSaved: () => void }) {
  const toast = useToast();
  const [date, setDate] = useState(toDateKey(addDays(new Date(), 1)));
  const [from, setFrom] = useState('09:00');
  const [to, setTo] = useState('13:00');
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const save = async () => {
    setError(null);
    if (from >= to) { setError('Окончание должно быть позже начала'); return; }
    if (new Date(localIso(date, from)).getTime() < Date.now()) { setError('Нельзя открыть время в прошлом'); return; }
    setSaving(true);
    try {
      await instructorApi.addAvailability({ start: localIso(date, from), end: localIso(date, to) });
      toast('Свободное время добавлено');
      onSaved();
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Ошибка сохранения');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Добавить свободное время">
      <div className="form">
        <label className="field"><span className="field__label">Дата</span>
          <input className="input" type="date" value={date} min={toDateKey(new Date())} onChange={(e) => setDate(e.target.value)} />
        </label>
        <div className="grid grid--2 grid--tight">
          <label className="field"><span className="field__label">С</span>
            <input className="input" type="time" step={3600} value={from} onChange={(e) => setFrom(e.target.value)} />
          </label>
          <label className="field"><span className="field__label">До</span>
            <input className="input" type="time" step={3600} value={to} onChange={(e) => setTo(e.target.value)} />
          </label>
        </div>
        {error && <Alert>{error}</Alert>}
        <Button block loading={saving} onClick={save}>Сохранить</Button>
      </div>
    </Modal>
  );
}

// =================== Экран 10: профиль инструктора ===================
export function InstructorProfilePage() {
  const me = useAsync(() => instructorApi.me(), []);
  return (
    <>
      <PageHead title="Информация об инструкторе" text="Просматривай свои личные данные и информацию о специализации." />
      <AsyncView state={me}>
        {(i) => (
          <div className="card panel instructor-profile">
            <Photo src={i.photoUrl} alt={i.name} label="Фото инструктора" className="instructor-profile__photo" />
            <div className="instructor-profile__info">
              <h2 className="h2">{i.name}</h2>
              <dl className="info-list">
                <div><dt>Специализация</dt><dd>{i.specialization}</dd></div>
                <div><dt>Стаж</dt><dd>{i.experienceYears} лет</dd></div>
                <div><dt>Email</dt><dd>{i.email}</dd></div>
                <div><dt>Стоимость часа</dt><dd>{i.pricePerHour.toLocaleString('ru-RU')} руб.</dd></div>
              </dl>
              <h4 className="list-heading">О себе</h4>
              <p>{i.bio || '—'}</p>
              <p className="muted small">Специализацию и стаж изменяет администратор.</p>
            </div>
          </div>
        )}
      </AsyncView>
      <div className="grid grid--2 section--gap">
        <ProfileForm />
        <PasswordForm />
      </div>
    </>
  );
}
