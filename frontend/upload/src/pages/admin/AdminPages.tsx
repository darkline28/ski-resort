import { useState } from 'react';
import { adminApi } from '@/api';
import { ApiError } from '@/api/client';
import { useAsync } from '@/hooks/useAsync';
import { isEmail, useForm } from '@/hooks/useForm';
import { useToast } from '@/context/ToastContext';
import { PageHead } from '@/layouts/CabinetLayout';
import { AsyncView, EmptyState } from '@/components/States';
import { Alert, Button, Chips, Field, Modal, StatusBadge, TextArea } from '@/components/ui';
import type { BookingStatus, Equipment, EquipmentCategory, Lesson, LessonCategory, Role, User } from '@/types';
import { fmtShort } from '@/utils/date';
import { EQUIPMENT_CATEGORY_LABEL, fmtPrice, LESSON_CATEGORY_LABEL, ROLE_LABEL, STATUS_LABEL } from '@/utils/format';

const WEEKDAYS = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];

// =================== Экран 11: главная админа ===================
export function AdminHomePage() {
  const stats = useAsync(() => adminApi.stats(), []);
  return (
    <>
      <PageHead title="Панель управления курортом" text="Здесь собрана основная информация о работе сайта, пользователях и бронированиях." />
      <AsyncView state={stats}>
        {(s) => {
          const max = Math.max(1, ...s.bookingsByWeekday);
          const top = s.bookingsByWeekday.indexOf(max);
          return (
            <>
              <div className="stats">
                <Stat label="Клиенты" value={s.users} />
                <Stat label="Инструкторы" value={s.instructors} />
                <Stat label="Брони сегодня" value={s.bookingsToday} />
                <Stat label="Выручка за неделю" value={s.revenueWeek >= 1000 ? `${Math.round(s.revenueWeek / 1000)} тыс.` : s.revenueWeek} />
              </div>
              <div className="grid grid--2 section--gap">
                <div className="card panel">
                  <h3 className="h3">Брони по дням недели</h3>
                  <div className="bars" role="img" aria-label="Количество броней по дням недели">
                    {s.bookingsByWeekday.map((v, i) => (
                      <div key={i} className="bars__col">
                        <span className="bars__value">{v}</span>
                        <div className={`bars__bar ${i === top ? 'bars__bar--top' : ''}`} style={{ height: `${Math.max(12, (v / max) * 100)}%` }}>
                          <span>{WEEKDAYS[i]}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="card panel">
                  <h3 className="h3">Последние брони</h3>
                  {s.latestBookings.length === 0 ? <p className="muted">Броней пока нет.</p> : (
                    <div className="list">
                      {s.latestBookings.map((b) => (
                        <div key={b.id} className="booking-row">
                          <div className="booking-row__time">№ {b.id}</div>
                          <div className="booking-row__info">
                            <div className="booking-row__title">{b.clientName}</div>
                            <div className="booking-row__sub">{b.title}</div>
                          </div>
                          <div className="booking-row__side"><StatusBadge status={b.status} /></div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </>
          );
        }}
      </AsyncView>
    </>
  );
}

function Stat({ label, value }: { label: string; value: number | string }) {
  return (
    <div className="card stat">
      <div className="stat__label">{label}</div>
      <div className="stat__value">{value}</div>
    </div>
  );
}

// =================== Экран 12: пользователи ===================
export function AdminUsersPage() {
  const [role, setRole] = useState<Role | 'all'>('all');
  const users = useAsync(() => adminApi.users(role === 'all' ? undefined : role), [role]);
  const [editing, setEditing] = useState<User | null>(null);
  const toast = useToast();

  const form = useForm({ name: '', email: '', password: '', specialization: '' }, (v) => ({
    name: v.name.trim().split(/\s+/).length >= 2 ? undefined : 'Укажи имя и фамилию',
    email: isEmail(v.email) ? undefined : 'Введите корректный email',
    password: v.password.length >= 6 ? undefined : 'Минимум 6 символов',
    specialization: v.specialization.trim() ? undefined : 'Укажи специализацию',
  }));

  return (
    <>
      <PageHead title="Управление пользователями" text="Просматривай список пользователей и управляй аккаунтами инструкторов." />
      <div className="admin-layout">
        <div className="card panel">
          <Chips options={[{ value: 'all', label: 'Все' }, { value: 'client', label: 'Клиенты' }, { value: 'instructor', label: 'Инструкторы' }, { value: 'admin', label: 'Админы' }]}
            value={role} onChange={setRole} />
          <AsyncView state={users} empty={<EmptyState text="Пользователей с такой ролью нет." />}>
            {(list) => (
              <div className="table-wrap">
                <table className="table">
                  <thead><tr><th>Имя</th><th>Email</th><th>Роль</th><th /></tr></thead>
                  <tbody>
                    {list.map((u) => (
                      <tr key={u.id}>
                        <td>{u.name}</td>
                        <td>{u.email}</td>
                        <td><span className="pill">{ROLE_LABEL[u.role]}</span></td>
                        <td className="table__actions"><Button variant="outline" className="btn--sm" onClick={() => setEditing(u)}>Изменить</Button></td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </AsyncView>
        </div>

        <form className="card panel form" noValidate onSubmit={form.handleSubmit(async (v) => {
          await adminApi.createInstructor(v);
          toast(`Аккаунт инструктора ${v.name} создан`);
          form.reset();
          users.reload();
        })}>
          <h3 className="h3">Новый инструктор</h3>
          {form.formError && <Alert>{form.formError}</Alert>}
          <Field name="name" placeholder="Имя и фамилия" value={form.values.name} onChange={form.onChange} error={form.errors.name} />
          <Field name="email" type="email" placeholder="Email" value={form.values.email} onChange={form.onChange} error={form.errors.email} />
          <Field name="password" placeholder="Временный пароль" value={form.values.password} onChange={form.onChange} error={form.errors.password} />
          <Field name="specialization" placeholder="Специализация" value={form.values.specialization} onChange={form.onChange} error={form.errors.specialization} />
          <Button type="submit" block loading={form.submitting}>Создать аккаунт</Button>
        </form>
      </div>

      {editing && (
        <EditUserModal user={editing} onClose={() => setEditing(null)} onSaved={(u) => {
          users.setData((list) => (list ?? []).map((x) => (x.id === u.id ? u : x)));
          setEditing(null);
          toast('Пользователь обновлён');
        }} />
      )}
    </>
  );
}

function EditUserModal({ user, onClose, onSaved }: { user: User; onClose: () => void; onSaved: (u: User) => void }) {
  const form = useForm({ name: user.name, email: user.email, phone: user.phone }, (v) => ({
    name: v.name.trim() ? undefined : 'Укажи имя',
    email: isEmail(v.email) ? undefined : 'Введите корректный email',
  }));
  return (
    <Modal open onClose={onClose} title="Редактирование пользователя">
      <form className="form" noValidate onSubmit={form.handleSubmit(async (v) => onSaved(await adminApi.updateUser(user.id, v)))}>
        {form.formError && <Alert>{form.formError}</Alert>}
        <Field label="Имя" name="name" value={form.values.name} onChange={form.onChange} error={form.errors.name} />
        <Field label="Email" name="email" value={form.values.email} onChange={form.onChange} error={form.errors.email} />
        <Field label="Телефон" name="phone" value={form.values.phone} onChange={form.onChange} />
        <Button type="submit" block loading={form.submitting}>Сохранить</Button>
      </form>
    </Modal>
  );
}

// =================== Экран 13: каталог ===================
type CatalogTab = 'lessons' | 'equipment';

export function AdminCatalogPage() {
  const [tab, setTab] = useState<CatalogTab>('lessons');
  const lessons = useAsync(() => adminApi.lessons(), []);
  const equipment = useAsync(() => adminApi.equipment(), []);
  const [editing, setEditing] = useState<Lesson | Equipment | 'new' | null>(null);
  const toast = useToast();

  const remove = async (item: Lesson | Equipment) => {
    if (!confirm(`Удалить «${item.title}»?`)) return;
    try {
      if (tab === 'lessons') { await adminApi.deleteLesson(item.id); lessons.reload(); }
      else { await adminApi.deleteEquipment(item.id); equipment.reload(); }
      toast('Удалено');
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Не удалось удалить', 'error');
    }
  };

  const state = tab === 'lessons' ? lessons : equipment;

  return (
    <>
      <PageHead title="Управление каталогом" text="Добавляй, редактируй и удаляй занятия и оборудование." />
      <div className="card panel">
        <div className="panel__head">
          <Chips options={[{ value: 'lessons', label: 'Занятия' }, { value: 'equipment', label: 'Оборудование' }]} value={tab} onChange={setTab} />
          <Button onClick={() => setEditing('new')}>Добавить</Button>
        </div>
        <AsyncView state={state as typeof lessons} empty={<EmptyState text="Каталог пуст. Добавь первую позицию." />}>
          {(list) => (
            <div className="table-wrap">
              <table className="table">
                <thead><tr><th>Название</th><th>Категория</th><th>Цена</th><th>Статус</th><th /></tr></thead>
                <tbody>
                  {(list as (Lesson | Equipment)[]).map((item) => (
                    <tr key={item.id}>
                      <td>{item.title}</td>
                      <td className="muted">{'price' in item ? LESSON_CATEGORY_LABEL[item.category] : EQUIPMENT_CATEGORY_LABEL[item.category]}</td>
                      <td>{'price' in item ? fmtPrice(item.price) : `${item.pricePerHour} руб./час`}</td>
                      <td><span className={`pill ${item.active ? 'pill--ok' : ''}`}>{item.active ? 'Активно' : 'Скрыто'}</span></td>
                      <td className="table__actions">
                        <Button variant="outline" className="btn--sm" onClick={() => setEditing(item)}>Изменить</Button>
                        <Button variant="danger" className="btn--sm" onClick={() => remove(item)}>Удалить</Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </AsyncView>
      </div>

      {editing && (tab === 'lessons'
        ? <LessonModal item={editing === 'new' ? null : (editing as Lesson)} onClose={() => setEditing(null)} onSaved={() => { setEditing(null); lessons.reload(); toast('Сохранено'); }} />
        : <EquipmentModal item={editing === 'new' ? null : (editing as Equipment)} onClose={() => setEditing(null)} onSaved={() => { setEditing(null); equipment.reload(); toast('Сохранено'); }} />)}
    </>
  );
}

const numberError = (v: string) => (Number(v) > 0 ? undefined : 'Положительное число');

function LessonModal({ item, onClose, onSaved }: { item: Lesson | null; onClose: () => void; onSaved: () => void }) {
  const form = useForm({
    title: item?.title ?? '', description: item?.description ?? '', category: item?.category ?? 'ski',
    price: String(item?.price ?? ''), durationMin: String(item?.durationMin ?? 60), active: item?.active === false ? 'false' : 'true',
  }, (v) => ({ title: v.title.trim() ? undefined : 'Укажи название', price: numberError(v.price), durationMin: numberError(v.durationMin) }));

  return (
    <Modal open onClose={onClose} title={item ? 'Редактировать занятие' : 'Новое занятие'}>
      <form className="form" noValidate onSubmit={form.handleSubmit(async (v) => {
        const payload = { title: v.title, description: v.description, category: v.category as LessonCategory, price: Number(v.price), durationMin: Number(v.durationMin), active: v.active === 'true' };
        if (item) await adminApi.updateLesson(item.id, payload); else await adminApi.createLesson(payload);
        onSaved();
      })}>
        {form.formError && <Alert>{form.formError}</Alert>}
        <Field label="Название" name="title" value={form.values.title} onChange={form.onChange} error={form.errors.title} />
        <TextArea label="Описание" name="description" rows={3} value={form.values.description} onChange={form.onChange} />
        <label className="field"><span className="field__label">Категория</span>
          <select className="input" name="category" value={form.values.category} onChange={form.onChange}>
            {Object.entries(LESSON_CATEGORY_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </label>
        <div className="grid grid--2 grid--tight">
          <Field label="Цена, руб." name="price" type="number" min={0} value={form.values.price} onChange={form.onChange} error={form.errors.price} />
          <Field label="Длительность, мин" name="durationMin" type="number" min={30} step={30} value={form.values.durationMin} onChange={form.onChange} error={form.errors.durationMin} />
        </div>
        <label className="field"><span className="field__label">Статус</span>
          <select className="input" name="active" value={form.values.active} onChange={form.onChange}>
            <option value="true">Активно</option><option value="false">Скрыто</option>
          </select>
        </label>
        <Button type="submit" block loading={form.submitting}>Сохранить</Button>
      </form>
    </Modal>
  );
}

function EquipmentModal({ item, onClose, onSaved }: { item: Equipment | null; onClose: () => void; onSaved: () => void }) {
  const form = useForm({
    title: item?.title ?? '', description: item?.description ?? '', category: item?.category ?? 'ski',
    pricePerHour: String(item?.pricePerHour ?? ''), sizes: item?.sizes.join(', ') ?? '', stock: String(item?.stock ?? 10),
    active: item?.active === false ? 'false' : 'true',
  }, (v) => ({
    title: v.title.trim() ? undefined : 'Укажи название', pricePerHour: numberError(v.pricePerHour), stock: numberError(v.stock),
    sizes: v.sizes.trim() ? undefined : 'Хотя бы один размер',
  }));

  return (
    <Modal open onClose={onClose} title={item ? 'Редактировать снаряжение' : 'Новое снаряжение'}>
      <form className="form" noValidate onSubmit={form.handleSubmit(async (v) => {
        const payload = {
          title: v.title, description: v.description, category: v.category as EquipmentCategory, pricePerHour: Number(v.pricePerHour),
          sizes: v.sizes.split(',').map((s) => s.trim()).filter(Boolean), stock: Number(v.stock), active: v.active === 'true',
        };
        if (item) await adminApi.updateEquipment(item.id, payload); else await adminApi.createEquipment(payload);
        onSaved();
      })}>
        {form.formError && <Alert>{form.formError}</Alert>}
        <Field label="Название" name="title" value={form.values.title} onChange={form.onChange} error={form.errors.title} />
        <TextArea label="Описание" name="description" rows={2} value={form.values.description} onChange={form.onChange} />
        <label className="field"><span className="field__label">Категория</span>
          <select className="input" name="category" value={form.values.category} onChange={form.onChange}>
            {Object.entries(EQUIPMENT_CATEGORY_LABEL).map(([k, l]) => <option key={k} value={k}>{l}</option>)}
          </select>
        </label>
        <div className="grid grid--2 grid--tight">
          <Field label="Цена, руб./час" name="pricePerHour" type="number" min={0} value={form.values.pricePerHour} onChange={form.onChange} error={form.errors.pricePerHour} />
          <Field label="Количество на складе" name="stock" type="number" min={0} value={form.values.stock} onChange={form.onChange} error={form.errors.stock} />
        </div>
        <Field label="Размеры (через запятую)" name="sizes" value={form.values.sizes} onChange={form.onChange} error={form.errors.sizes} />
        <label className="field"><span className="field__label">Статус</span>
          <select className="input" name="active" value={form.values.active} onChange={form.onChange}>
            <option value="true">Активно</option><option value="false">Скрыто</option>
          </select>
        </label>
        <Button type="submit" block loading={form.submitting}>Сохранить</Button>
      </form>
    </Modal>
  );
}

// =================== Экран 14: бронирования ===================
const STATUSES: BookingStatus[] = ['pending', 'paid', 'cancelled', 'completed'];

export function AdminBookingsPage() {
  const [status, setStatus] = useState<BookingStatus | 'all'>('all');
  const bookings = useAsync(() => adminApi.bookings(status === 'all' ? undefined : status), [status]);
  const toast = useToast();

  const change = async (id: number, next: BookingStatus) => {
    try {
      const updated = await adminApi.setBookingStatus(id, next);
      bookings.setData((list) => (list ?? [])
        .map((b) => (b.id === id ? updated : b))
        .filter((b) => status === 'all' || b.status === status));
      toast(`Статус брони № ${id}: ${STATUS_LABEL[next]}`);
    } catch (e) {
      toast(e instanceof ApiError ? e.message : 'Не удалось изменить статус', 'error');
    }
  };

  return (
    <>
      <PageHead title="Все бронирования" text="Просматривай заказы клиентов, проверяй их статус и управляй бронированиями." />
      <div className="card panel">
        <Chips options={[{ value: 'all', label: 'Все' }, ...STATUSES.map((s) => ({ value: s, label: STATUS_LABEL[s] }))]} value={status} onChange={setStatus} />
        <AsyncView state={bookings} empty={<EmptyState text="Броней с таким статусом нет." />}>
          {(list) => (
            <div className="table-wrap">
              <table className="table">
                <thead><tr><th>№</th><th>Клиент</th><th>Заказ</th><th>Дата</th><th>Сумма</th><th>Статус</th></tr></thead>
                <tbody>
                  {list.map((b) => (
                    <tr key={b.id}>
                      <td>{b.id}</td>
                      <td>{b.clientName}</td>
                      <td>{b.title}{b.instructorName ? <div className="muted small">{b.instructorName}</div> : null}</td>
                      <td>{fmtShort(b.start)}</td>
                      <td>{fmtPrice(b.total)}</td>
                      <td>
                        <select className={`status-select badge badge--${b.status}`} value={b.status} onChange={(e) => change(b.id, e.target.value as BookingStatus)} aria-label={`Статус брони ${b.id}`}>
                          {STATUSES.map((s) => <option key={s} value={s}>{STATUS_LABEL[s]}</option>)}
                        </select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </AsyncView>
      </div>
    </>
  );
}
