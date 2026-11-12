import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { cabinetPath } from '@/utils/routes';
import { useToast } from '@/context/ToastContext';
import { isEmail, isPhone, useForm } from '@/hooks/useForm';
import { Alert, Button, Field } from '@/components/ui';
import { USE_MOCKS } from '@/api/client';
import { useEffect, useRef } from 'react';

/** Экран 4: вход и регистрация на одной странице. /login и /register скроллят к нужной форме */
export default function AuthPage({ mode }: { mode: 'login' | 'register' }) {
  const { login, register, user } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from;
  const registerRef = useRef<HTMLFormElement>(null);

  useEffect(() => {
    if (mode === 'register') registerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, [mode]);

  useEffect(() => {
    if (user) navigate(from ?? cabinetPath(user), { replace: true });
  }, [user]); // eslint-disable-line react-hooks/exhaustive-deps

  const loginForm = useForm({ email: '', password: '' }, (v) => ({
    email: isEmail(v.email) ? undefined : 'Введите корректный email',
    password: v.password.length >= 6 ? undefined : 'Пароль — минимум 6 символов',
  }));

  const regForm = useForm({ name: '', phone: '', email: '', password: '' }, (v) => ({
    name: v.name.trim().length >= 2 ? undefined : 'Укажи имя',
    phone: isPhone(v.phone) ? undefined : 'Укажи телефон (минимум 10 цифр)',
    email: isEmail(v.email) ? undefined : 'Введите корректный email',
    password: v.password.length >= 6 ? undefined : 'Пароль — минимум 6 символов',
  }));

  return (
    <section className="container section section--top">
      <div className="grid grid--2">
        <form className="card form auth-card" noValidate onSubmit={loginForm.handleSubmit(async (v) => {
          const u = await login(v);
          toast('Добро пожаловать! Вы успешно вошли.');
          navigate(from ?? cabinetPath(u), { replace: true });
        })}>
          <h1 className="h2">С возвращением!</h1>
          <p className="lead">Войди в личный кабинет, чтобы управлять своими бронированиями и просматривать информацию о занятиях.</p>
          {loginForm.formError && <Alert>{loginForm.formError}</Alert>}
          <Field name="email" type="email" autoComplete="email" placeholder="Email" value={loginForm.values.email} onChange={loginForm.onChange} error={loginForm.errors.email} />
          <Field name="password" type="password" autoComplete="current-password" placeholder="Пароль" value={loginForm.values.password} onChange={loginForm.onChange} error={loginForm.errors.password} />
          <Button type="submit" block loading={loginForm.submitting}>Войти</Button>
          <p className="form__switch">Нет аккаунта? <button type="button" className="link" onClick={() => navigate('/register')}>Зарегистрироваться</button></p>
          {USE_MOCKS && (
            <div className="demo-hint">
              <b>Демо-аккаунты (заглушка API):</b><br />
              клиент — anna@example.com / 123456<br />
              инструктор — alexey@montis.ru / 123456<br />
              админ — admin@montis.ru / admin123
            </div>
          )}
        </form>

        <form ref={registerRef} className="card form auth-card" noValidate onSubmit={regForm.handleSubmit(async (v) => {
          const u = await register(v);
          toast('Аккаунт создан! Добро пожаловать в Montis.');
          navigate(from ?? cabinetPath(u), { replace: true });
        })}>
          <h2 className="h2">Создай свой аккаунт</h2>
          <p className="lead">Зарегистрируйся, чтобы бронировать занятия, арендовать оборудование и следить за своими заказами.</p>
          {regForm.formError && <Alert>{regForm.formError}</Alert>}
          <Field name="name" autoComplete="name" placeholder="Имя" value={regForm.values.name} onChange={regForm.onChange} error={regForm.errors.name} />
          <Field name="phone" type="tel" autoComplete="tel" placeholder="Телефон" value={regForm.values.phone} onChange={regForm.onChange} error={regForm.errors.phone} />
          <Field name="email" type="email" autoComplete="email" placeholder="Email" value={regForm.values.email} onChange={regForm.onChange} error={regForm.errors.email} />
          <Field name="password" type="password" autoComplete="new-password" placeholder="Пароль" value={regForm.values.password} onChange={regForm.onChange} error={regForm.errors.password} />
          <Button type="submit" block loading={regForm.submitting}>Зарегистрироваться</Button>
          <p className="form__switch">Уже есть аккаунт? <button type="button" className="link" onClick={() => navigate('/login')}>Войти</button></p>
          <p className="form__note">Аккаунты инструкторов создаёт администратор.</p>
        </form>
      </div>
    </section>
  );
}
