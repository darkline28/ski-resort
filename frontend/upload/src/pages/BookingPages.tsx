import { useEffect, useMemo, useState } from 'react';
import { Link, useLocation, useNavigate, useParams, useSearchParams } from 'react-router-dom';
import { bookingApi, catalogApi } from '@/api';
import { ApiError } from '@/api/client';
import { useAsync } from '@/hooks/useAsync';
import { useAuth } from '@/context/AuthContext';
import { useToast } from '@/context/ToastContext';
import { AsyncView, ErrorState, Loader } from '@/components/States';
import { Alert, Button, Chips } from '@/components/ui';
import { addDays, fmtDay, fmtDayLong, fmtDuration, fmtTime, fmtWeekdayShort, startOfDay, toDateKey, toInputDateTime } from '@/utils/date';
import { fmtPrice, shortName } from '@/utils/format';

/** Общая проверка: бронировать может только клиент; гостя отправляем на вход */
function useRequireClient() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const toast = useToast();
  return (): boolean => {
    if (!user) {
      toast('Войди или зарегистрируйся, чтобы оформить бронь', 'error');
      navigate('/login', { state: { from: location.pathname + location.search } });
      return false;
    }
    if (user.role !== 'client') {
      toast('Бронировать может только клиент', 'error');
      return false;
    }
    return true;
  };
}

// =================== Экран 15: бронирование занятия ===================
export function BookLessonPage() {
  const { id } = useParams();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const toast = useToast();
  const requireClient = useRequireClient();

  const lessons = useAsync(() => catalogApi.lessons(), []);
  const [lessonId, setLessonId] = useState<number | null>(id ? Number(id) : null);
  const lesson = lessons.data?.find((l) => l.id === lessonId) ?? null;

  useEffect(() => {
    if (!lessonId && lessons.data?.length) setLessonId(lessons.data[0].id);
  }, [lessons.data, lessonId]);

  const instructors = useAsync(
    () => (lessonId ? catalogApi.instructors(lessonId) : Promise.resolve([])),
    [lessonId],
  );
  const preferred = params.get('instructor') ? Number(params.get('instructor')) : null;
  const [instructorId, setInstructorId] = useState<number | null>(null);

  useEffect(() => {
    const list = instructors.data;
    if (!list) return;
    if (instructorId && list.some((i) => i.id === instructorId)) return;
    setInstructorId(list.find((i) => i.id === preferred)?.id ?? list[0]?.id ?? null);
  }, [instructors.data]); // eslint-disable-line react-hooks/exhaustive-deps

  const days = useMemo(() => Array.from({ length: 14 }, (_, i) => addDays(startOfDay(new Date()), i)), []);
  const [date, setDate] = useState(toDateKey(days[0]));
  const [slotStart, setSlotStart] = useState<string | null>(null);

  // Свободное время приходит с бэкенда — фронт его не вычисляет
  const slots = useAsync(
    () => (instructorId && lessonId ? catalogApi.slots(instructorId, date, lessonId) : Promise.resolve([])),
    [instructorId, date, lessonId],
  );
  useEffect(() => setSlotStart(null), [instructorId, date, lessonId]);

  const slot = slots.data?.find((s) => s.start === slotStart) ?? null;
  const instructor = instructors.data?.find((i) => i.id === instructorId) ?? null;
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (!lesson || !instructor || !slot) return;
    if (!requireClient()) return;
    setSubmitting(true);
    setError(null);
    try {
      const booking = await bookingApi.createLesson({ lessonId: lesson.id, instructorId: instructor.id, start: slot.start });
      toast('Запись создана. Осталось оплатить.');
      navigate(`/checkout/${booking.id}`);
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Не удалось создать запись.';
      setError(msg);
      if (e instanceof ApiError && e.status === 409) slots.reload(); // интервал заняли — обновим сетку
    } finally {
      setSubmitting(false);
    }
  };

  if (lessons.error) return <div className="container section"><ErrorState message={lessons.error.message} onRetry={lessons.reload} /></div>;

  return (
    <section className="container section section--top">
      <h1 className="h1">Забронируй своё занятие</h1>
      <p className="lead">Выбери инструктора, удобную дату и время. Осталось совсем немного до твоего первого спуска!</p>

      <div className="booking">
        <div className="card booking__form">
          <div className="step">
            <div className="step__label">1. Занятие</div>
            {lessons.loading ? <Loader /> : (
              <Chips options={(lessons.data ?? []).map((l) => ({ value: l.id, label: l.title }))} value={lessonId} onChange={setLessonId} />
            )}
          </div>

          <div className="step">
            <div className="step__label">2. Инструктор</div>
            <AsyncView state={instructors} loading={<Loader />}
              empty={<p className="muted">Нет инструкторов для этого занятия.</p>}>
              {(list) => <Chips options={list.map((i) => ({ value: i.id, label: shortName(i.name) }))} value={instructorId} onChange={setInstructorId} />}
            </AsyncView>
          </div>

          <div className="step">
            <div className="step__label">3. Дата</div>
            <div className="scroll-x">
              <Chips
                options={days.map((d) => ({ value: toDateKey(d), label: `${fmtWeekdayShort(d)} ${d.getDate()}` }))}
                value={date}
                onChange={setDate}
              />
            </div>
          </div>

          <div className="step">
            <div className="step__label">4. Время</div>
            {!instructorId ? <p className="muted">Сначала выбери инструктора.</p> : (
              <AsyncView state={slots} loading={<Loader text="Получаем свободное время" />}
                empty={<p className="muted">На эту дату нет свободного времени. Выбери другой день.</p>}>
                {(list) => (
                  <Chips
                    options={list.map((s) => ({ value: s.start, label: fmtTime(s.start), disabled: !s.available }))}
                    value={slotStart}
                    onChange={setSlotStart}
                  />
                )}
              </AsyncView>
            )}
          </div>
        </div>

        <aside className="summary">
          <h3 className="summary__title">Ваша запись</h3>
          <dl className="summary__rows">
            <Row label="Занятие" value={lesson?.title} />
            <Row label="Инструктор" value={instructor ? shortName(instructor.name) : undefined} />
            <Row label="Дата" value={fmtDayLong(date + 'T12:00:00')} />
            <Row label="Время" value={slot ? `${fmtTime(slot.start)}–${fmtTime(slot.end)}` : undefined} />
            {lesson && <Row label="Длительность" value={fmtDuration(lesson.durationMin)} />}
          </dl>
          <div className="summary__total"><span>Итого</span><strong>{lesson ? fmtPrice(lesson.price) : '—'}</strong></div>
          {error && <Alert>{error}</Alert>}
          <Button variant="light" block disabled={!slot} loading={submitting} onClick={submit}>Продолжить</Button>
        </aside>
      </div>
    </section>
  );
}

function Row({ label, value }: { label: string; value?: string }) {
  return (
    <div className="summary__row">
      <dt>{label}</dt>
      <dd>{value ?? '—'}</dd>
    </div>
  );
}

// =================== Экран 16: аренда оборудования ===================
export function BookRentalPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const toast = useToast();
  const requireClient = useRequireClient();

  const equipment = useAsync(() => catalogApi.equipment(), []);
  const [itemId, setItemId] = useState<number | null>(id ? Number(id) : null);
  const item = equipment.data?.find((e) => e.id === itemId) ?? null;
  useEffect(() => {
    if (!itemId && equipment.data?.length) setItemId(equipment.data[0].id);
  }, [equipment.data, itemId]);

  const [size, setSize] = useState<string | null>(null);
  useEffect(() => setSize(item?.sizes[Math.floor(item.sizes.length / 2)] ?? null), [item]);
  const [quantity, setQuantity] = useState(1);

  const defaultStart = useMemo(() => {
    const d = addDays(startOfDay(new Date()), 1);
    d.setHours(10);
    return d;
  }, []);
  const [start, setStart] = useState(toInputDateTime(defaultStart));
  const [end, setEnd] = useState(toInputDateTime(new Date(defaultStart.getTime() + 4 * 3600000)));

  const startIso = start ? new Date(start).toISOString() : '';
  const endIso = end ? new Date(end).toISOString() : '';
  const hours = start && end ? Math.ceil((new Date(end).getTime() - new Date(start).getTime()) / 3600000) : 0;
  const periodError =
    !start || !end ? 'Укажи период аренды'
      : hours < 1 ? 'Окончание должно быть минимум через час после начала'
        : new Date(start).getTime() < Date.now() ? 'Начало аренды уже прошло' : null;

  // Остаток на складе на выбранный период считает бэкенд
  const availability = useAsync(
    () => (item && size && !periodError ? catalogApi.equipmentAvailability(item.id, startIso, endIso, size) : Promise.resolve(null)),
    [item?.id, size, startIso, endIso, periodError],
  );
  const maxQty = availability.data?.available ?? item?.stock ?? 1;

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const submit = async () => {
    if (!item || !size || periodError) return;
    if (!requireClient()) return;
    setSubmitting(true);
    setError(null);
    try {
      const b = await bookingApi.createRental({ equipmentId: item.id, size, quantity, start: startIso, end: endIso });
      toast('Заказ создан. Осталось оплатить.');
      navigate(`/checkout/${b.id}`);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'Не удалось оформить аренду.');
      availability.reload();
    } finally {
      setSubmitting(false);
    }
  };

  if (equipment.error) return <div className="container section"><ErrorState message={equipment.error.message} onRetry={equipment.reload} /></div>;

  return (
    <section className="container section section--top">
      <h1 className="h1">Выбери необходимое снаряжение</h1>
      <p className="lead">Укажи размер, количество и период аренды, чтобы подготовиться к катанию.</p>

      <div className="booking">
        <div className="card booking__form">
          <div className="step">
            <div className="step__label">Снаряжение</div>
            {equipment.loading ? <Loader /> : (
              <Chips options={(equipment.data ?? []).map((e) => ({ value: e.id, label: e.title }))} value={itemId} onChange={(v) => { setItemId(v); setQuantity(1); }} />
            )}
          </div>
          {item && (
            <div className="step">
              <div className="step__label">Размер</div>
              <Chips options={item.sizes.map((s) => ({ value: s, label: s }))} value={size} onChange={setSize} />
            </div>
          )}
          <div className="step">
            <div className="step__label">Количество</div>
            <div className="stepper">
              <button type="button" onClick={() => setQuantity((q) => Math.max(1, q - 1))} disabled={quantity <= 1} aria-label="Меньше">−</button>
              <span aria-live="polite">{quantity}</span>
              <button type="button" onClick={() => setQuantity((q) => Math.min(maxQty, q + 1))} disabled={quantity >= maxQty} aria-label="Больше">+</button>
            </div>
            {availability.data && <p className="muted small">Доступно на период: {availability.data.available} шт.</p>}
          </div>
          <div className="step">
            <div className="step__label">Период аренды</div>
            <div className="grid grid--2 grid--tight">
              <label className="field">
                <span className="field__label">Начало</span>
                <input className="input" type="datetime-local" step={3600} value={start} onChange={(e) => setStart(e.target.value)} />
              </label>
              <label className="field">
                <span className="field__label">Окончание</span>
                <input className="input" type="datetime-local" step={3600} value={end} onChange={(e) => setEnd(e.target.value)} />
              </label>
            </div>
            {periodError && <p className="field__error">{periodError}</p>}
          </div>
        </div>

        <aside className="summary">
          <h3 className="summary__title">Ваш заказ</h3>
          <dl className="summary__rows">
            <Row label={item ? `${item.title} × ${quantity}` : 'Снаряжение'} value={item ? `${item.pricePerHour.toLocaleString('ru-RU')} руб./час` : undefined} />
            <Row label="Размер" value={size ?? undefined} />
            <Row label="Период" value={!periodError ? `${fmtDay(startIso)}, ${fmtDuration(hours * 60)}` : undefined} />
          </dl>
          <div className="summary__total">
            <span>Итого</span>
            <strong>{item && !periodError ? fmtPrice(item.pricePerHour * hours * quantity) : '—'}</strong>
          </div>
          {quantity > maxQty && <Alert>На этот период доступно только {maxQty} шт.</Alert>}
          {error && <Alert>{error}</Alert>}
          <Button variant="light" block disabled={!item || !size || !!periodError || quantity > maxQty} loading={submitting} onClick={submit}>
            Продолжить
          </Button>
        </aside>
      </div>
    </section>
  );
}

// =================== Экран 17: оплата (демо) ===================
export function CheckoutPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const booking = useAsync(() => bookingApi.get(Number(id)), [id]);
  const [card, setCard] = useState({ cardNumber: '', expiry: '', cvc: '' });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);

  const formatCard = (v: string) => v.replace(/\D/g, '').slice(0, 16).replace(/(\d{4})(?=\d)/g, '$1 ');
  const formatExpiry = (v: string) => v.replace(/\D/g, '').slice(0, 4).replace(/(\d{2})(?=\d)/, '$1/');

  const validate = () => {
    const e: Record<string, string> = {};
    if (card.cardNumber.replace(/\s/g, '').length !== 16) e.cardNumber = 'Номер карты — 16 цифр';
    const [mm] = card.expiry.split('/');
    if (!/^\d{2}\/\d{2}$/.test(card.expiry) || Number(mm) < 1 || Number(mm) > 12) e.expiry = 'Формат ММ/ГГ';
    if (!/^\d{3}$/.test(card.cvc)) e.cvc = '3 цифры';
    setErrors(e);
    return !Object.keys(e).length;
  };

  const pay = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      await bookingApi.pay(Number(id), card);
      navigate(`/checkout/${id}/result?status=success`, { replace: true });
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'Ошибка оплаты';
      navigate(`/checkout/${id}/result?status=fail&message=${encodeURIComponent(msg)}`, { replace: true });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <section className="container section section--top">
      <h1 className="h1">Проверь свой заказ</h1>
      <p className="lead">Убедись, что все данные указаны верно, и перейди к демонстрационной оплате.</p>
      <AsyncView state={booking}>
        {(b) => b.status !== 'pending' ? (
          <div className="card state">
            <h3>Этот заказ уже {b.status === 'paid' ? 'оплачен' : b.status === 'cancelled' ? 'отменён' : 'завершён'}</h3>
            <Link className="btn btn--primary" to="/account/bookings">К моим броням</Link>
          </div>
        ) : (
          <div className="booking">
            <form className="card booking__form form" noValidate onSubmit={(e) => { e.preventDefault(); pay(); }}>
              <h3 className="h3">Данные карты (демо)</h3>
              <label className={`field ${errors.cardNumber ? 'field--error' : ''}`}>
                <span className="field__label">Номер карты</span>
                <input className="input" inputMode="numeric" autoComplete="off" placeholder="0000 0000 0000 0000"
                  value={card.cardNumber} onChange={(e) => setCard({ ...card, cardNumber: formatCard(e.target.value) })} />
                {errors.cardNumber && <span className="field__error">{errors.cardNumber}</span>}
              </label>
              <div className="grid grid--2 grid--tight">
                <label className={`field ${errors.expiry ? 'field--error' : ''}`}>
                  <span className="field__label">Срок</span>
                  <input className="input" inputMode="numeric" autoComplete="off" placeholder="ММ/ГГ"
                    value={card.expiry} onChange={(e) => setCard({ ...card, expiry: formatExpiry(e.target.value) })} />
                  {errors.expiry && <span className="field__error">{errors.expiry}</span>}
                </label>
                <label className={`field ${errors.cvc ? 'field--error' : ''}`}>
                  <span className="field__label">CVC</span>
                  <input className="input" inputMode="numeric" type="password" autoComplete="off" placeholder="•••"
                    value={card.cvc} onChange={(e) => setCard({ ...card, cvc: e.target.value.replace(/\D/g, '').slice(0, 3) })} />
                  {errors.cvc && <span className="field__error">{errors.cvc}</span>}
                </label>
              </div>
              <Alert kind="info">Это демонстрационная оплата: реальные деньги не списываются. Любая карта проходит, а карта, оканчивающаяся на 0000, — отклоняется.</Alert>
              <button type="submit" hidden />
            </form>
            <aside className="summary">
              <h3 className="summary__title">Ваш заказ</h3>
              <dl className="summary__rows">
                <Row label={b.instructorName ? `${b.title}, ${shortName(b.instructorName)}` : b.title} value={fmtPrice(b.total)} />
                <Row label={`${fmtDay(b.start)}, ${fmtTime(b.start)}`} value={b.quantity ? `× ${b.quantity}` : ''} />
                <Row label="Статус" value="Ожидание оплаты" />
              </dl>
              <div className="summary__total"><span>Итого</span><strong>{fmtPrice(b.total)}</strong></div>
              <Button variant="light" block loading={submitting} onClick={pay}>Оплатить</Button>
            </aside>
          </div>
        )}
      </AsyncView>
    </section>
  );
}

// =================== Экран 18: результат оплаты ===================
export function PaymentResultPage() {
  const { id } = useParams();
  const [params] = useSearchParams();
  const success = params.get('status') === 'success';

  return (
    <section className="container section section--top">
      <div className={`card result ${success ? 'result--success' : 'result--fail'}`}>
        <div className="result__icon">{success ? '✓' : '×'}</div>
        {success ? (
          <>
            <h1 className="h2">Заказ успешно оформлен!</h1>
            <p className="lead">Твоё бронирование подтверждено. Подробности можно посмотреть в личном кабинете.</p>
            <Link className="btn btn--primary" to="/account/bookings">Перейти в личный кабинет</Link>
          </>
        ) : (
          <>
            <h1 className="h2">Не удалось завершить оплату</h1>
            <p className="lead">{params.get('message') || 'Попробуй ещё раз или вернись к оформлению заказа.'}</p>
            <Link className="btn btn--primary" to={`/checkout/${id}`}>Повторить оплату</Link>
          </>
        )}
      </div>
    </section>
  );
}
