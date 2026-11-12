import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { Logo } from '@/components/Logo';
import { Button } from '@/components/ui';
import { initials, ROLE_LABEL } from '@/utils/format';
import type { Role } from '@/types';

const MENU: Record<Role, { to: string; label: string; end?: boolean }[]> = {
  client: [
    { to: '/account', label: 'Главная', end: true },
    { to: '/account/bookings', label: 'Мои брони' },
    { to: '/account/profile', label: 'Мой профиль' },
  ],
  instructor: [
    { to: '/instructor', label: 'Главная', end: true },
    { to: '/instructor/schedule', label: 'Расписание и занятия' },
    { to: '/instructor/profile', label: 'Мой профиль' },
  ],
  admin: [
    { to: '/admin', label: 'Главная', end: true },
    { to: '/admin/users', label: 'Пользователи' },
    { to: '/admin/catalog', label: 'Каталог' },
    { to: '/admin/bookings', label: 'Бронирования' },
  ],
};

export function CabinetLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  if (!user) return null;

  const onLogout = async () => {
    await logout();
    navigate('/');
  };

  return (
    <div className="page cabinet">
      <header className="header container">
        <div className="header__bar">
          <Logo />
          <div className="user-chip">
            <span className="avatar">{initials(user.name)}</span>
            <span className="user-chip__text">
              <span className="user-chip__name">{user.name}</span>
              <span className="user-chip__role">{ROLE_LABEL[user.role]}</span>
            </span>
            <Button variant="light" onClick={onLogout}>Выйти</Button>
          </div>
        </div>
      </header>
      <div className="container cabinet__body">
        <aside className="sidebar card">
          <nav aria-label="Меню кабинета">
            {MENU[user.role].map((m) => (
              <NavLink key={m.to} to={m.to} end={m.end} className={({ isActive }) => `sidebar__link ${isActive ? 'sidebar__link--active' : ''}`}>
                {m.label}
              </NavLink>
            ))}
            <hr className="sidebar__divider" />
            <NavLink to="/" className="sidebar__link">На сайт</NavLink>
            <button className="sidebar__link sidebar__logout" onClick={onLogout}>Выйти</button>
          </nav>
        </aside>
        <main className="cabinet__main"><Outlet /></main>
      </div>
    </div>
  );
}

export function PageHead({ title, text, children }: { title: string; text?: string; children?: React.ReactNode }) {
  return (
    <div className="page-head">
      <div>
        <h1 className="h1">{title}</h1>
        {text && <p className="lead">{text}</p>}
      </div>
      {children}
    </div>
  );
}
