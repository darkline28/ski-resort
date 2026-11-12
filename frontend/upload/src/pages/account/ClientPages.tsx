import { useState } from 'react';
import { bookingApi, profileApi } from '@/api';
import { ApiError } from '@/api/client';
import { useAsync } from '@/hooks/useAsync';
import { isEmail, isPhone, useForm } from '@/hooks/useForm';
import { useAuth } from '@/context/AuthContext';
import { useToast } from '@/context/ToastContext';
import { PageHead } from '@/layouts/CabinetLayout';
import { AsyncView, EmptyState } from '@/components/States';
import { BookingRow } from '@/components/BookingRow';
import { Alert, Button, ButtonLink, Chips, Field } from '@/components/ui';
import type { Booking, BookingType } from '@/types';

const isUpcoming = (b: Booking) => new Date(b.end).getTime() > Date.now() && (b.status === 'pending' || b.status === 'paid');

// =================== Экран 5: главная ЛК клиента ===================
export function ClientHomePage() {
  const { user } = useAuth();
  const bookings = useAsync(() => bookingApi.my(), []);

  return (
    <>
      <PageHead title={`Добро пожаловать${user ? ', ' + user.name.split(' ')[0] : ''}!`}
        text="Здесь ты можешь просматривать свои ближайшие бронирования и управлять отдыхом на курорте." />
      <div className="grid grid--2">
        <div className="card panel">
          <h3 className="h3">Ближайшие брони</h3>
          <AsyncView state={bookings}
            isEmpty={(list) => !list.some(isUpcoming)}
            empty={<EmptyState text="Как только появятся бронирования, они отобразятся здесь." action={{ to: '/lessons', label: 'Выбрать занятие' }} />}>
            {(list) => (
              <div className="list">
                {list.filter(isUpcoming).slice(0, 4).map((b) => <BookingRow key={b.id} booking={b} />)}
              </div>
            )}
          </AsyncView>
        </div>
        <div className="promo">
          <h3>Хочешь записаться ещё?</h3>
          <p>Выбери занятие или арендуй снаряжение.</p>
          <div className="promo__actions">
            <ButtonLink to="/lessons" variant="light">Выбрать занятие</ButtonLink>
            <ButtonLink to="/rental" variant="ghost">Прокат</ButtonLink>
          </div>
        </div>
      </div>
    </>
  );
}

// =================== Экран 6: мои брони ===================
export function ClientBookingsPage() {
  const [type, setType] = useState<BookingType>('lesson');
  const bookings = useAsync(() => bookingApi.my(type), [type]);
  const toast = useToast();
  const [busyId, setBusyId] = useState<number | null>(null);

  const cancel = async (b: Booking) => {
    if (!confirm(`Отменить «${b.title}»?`)) return;
    setBusyId(b.id);
    try {
      const updated = await bookingApi.cancel(b.id);
      bookings.setData((list) => (list ?? []).map((x) => (x.id === b.id ? updated : x)));
      toast('Бронирование отменено');
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Не удалось отменить', 'error');
    } finally {
      setBusyId(null);
    }
  };

  const sorted = (list: Booking[]) => {
    const upcoming = list.filter(isUpcoming).sort((a, b) => a.start.localeCompare(b.start));
    const history = list.filter((b) => !isUpcoming(b)).sort((a, b) => b.start.localeCompare(a.start));
    return { upcoming, history };
  };

  return (
    <>
      <PageHead title="Твои бронирования" text="В этом разделе собраны все твои занятия и заказы на аренду оборудования." />
      <div className="card panel">
        <Chips options={[{ value: 'lesson', label: 'Занятия' }, { value: 'rental', label: 'Аренда оборудования' }]} value={type} onChange={setType} />
        <AsyncView state={bookings}
          empty={<EmptyState text="Пока ничего нет. Как только появятся бронирования, они отобразятся здесь."
            action={type === 'lesson' ? { to: '/lessons', label: 'Выбрать занятие' } : { to: '/rental', label: 'Перейти к прокату' }} />}>
          {(list) => {
            const { upcoming, history } = sorted(list);
            const row = (b: Booking) => (
              <BookingRow key={b.id} booking={b} meta="price" actions={isUpcoming(b) && (
                <>
                  {b.status === 'pending' && <ButtonLink to={`/checkout/${b.id}`} variant="primary" className="btn--sm">Оплатить</ButtonLink>}
                  <Button variant="outline" className="btn--sm" loading={busyId === b.id} onClick={() => cancel(b)}>Отменить</Button>
                </>
              )} />
            );
            return (
              <>
                <h4 className="list-heading">Предстоящие</h4>
                {upcoming.length ? <div className="list">{upcoming.map(row)}</div> : <p className="muted">Нет предстоящих броней.</p>}
                <h4 className="list-heading">История</h4>
                {history.length ? <div className="list">{history.map(row)}</div> : <p className="muted">История пуста.</p>}
              </>
            );
          }}
        </AsyncView>
      </div>
    </>
  );
}

// =================== Экран 7: мой профиль ===================
export function ClientProfilePage() {
  return (
    <>
      <PageHead title="Личные данные" text="Управляй информацией своего аккаунта и при необходимости обновляй пароль." />
      <div className="grid grid--2">
        <ProfileForm />
        <PasswordForm />
      </div>
    </>
  );
}

export function ProfileForm() {
  const { user, setUser } = useAuth();
  const toast = useToast();
  const form = useForm({ name: user?.name ?? '', phone: user?.phone ?? '', email: user?.email ?? '' }, (v) => ({
    name: v.name.trim().length >= 2 ? undefined : 'Укажи имя',
    phone: isPhone(v.phone) ? undefined : 'Минимум 10 цифр',
    email: isEmail(v.email) ? undefined : 'Введите корректный email',
  }));
  return (
    <form className="card form panel" noValidate onSubmit={form.handleSubmit(async (v) => {
      const u = await profileApi.update(v);
      setUser(u);
      toast('Данные сохранены');
    })}>
      <h3 className="h3">Профиль</h3>
      {form.formError && <Alert>{form.formError}</Alert>}
      <Field label="Имя" name="name" value={form.values.name} onChange={form.onChange} error={form.errors.name} />
      <Field label="Телефон" name="phone" type="tel" value={form.values.phone} onChange={form.onChange} error={form.errors.phone} />
      <Field label="Email" name="email" type="email" value={form.values.email} onChange={form.onChange} error={form.errors.email} />
      <Button type="submit" loading={form.submitting}>Сохранить</Button>
    </form>
  );
}

export function PasswordForm() {
  const toast = useToast();
  const form = useForm({ currentPassword: '', newPassword: '', repeat: '' }, (v) => ({
    currentPassword: v.currentPassword ? undefined : 'Введи текущий пароль',
    newPassword: v.newPassword.length >= 6 ? undefined : 'Минимум 6 символов',
    repeat: v.repeat === v.newPassword ? undefined : 'Пароли не совпадают',
  }));
  return (
    <form className="card form panel" noValidate onSubmit={form.handleSubmit(async (v) => {
      await profileApi.changePassword({ currentPassword: v.currentPassword, newPassword: v.newPassword });
      form.reset();
      toast('Пароль изменён');
    })}>
      <h3 className="h3">Смена пароля</h3>
      {form.formError && <Alert>{form.formError}</Alert>}
      <Field label="Текущий пароль" name="currentPassword" type="password" autoComplete="current-password" placeholder="••••••••" value={form.values.currentPassword} onChange={form.onChange} error={form.errors.currentPassword} />
      <Field label="Новый пароль" name="newPassword" type="password" autoComplete="new-password" placeholder="Минимум 6 символов" value={form.values.newPassword} onChange={form.onChange} error={form.errors.newPassword} />
      <Field label="Повтор нового пароля" name="repeat" type="password" autoComplete="new-password" placeholder="Повторите пароль" value={form.values.repeat} onChange={form.onChange} error={form.errors.repeat} />
      <Button type="submit" loading={form.submitting}>Сменить пароль</Button>
    </form>
  );
}
