import { useEffect, useState } from 'react';
import { Link, NavLink, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';
import { cabinetPath } from '@/utils/routes';
import { ButtonLink } from '@/components/ui';
import { Logo } from '@/components/Logo';
import { CONTACTS } from '@/config';

const NAV = [
  { to: '/lessons', label: 'Занятия' },
  { to: '/rental', label: 'Прокат' },
  { to: '/#instructors', label: 'Инструкторы' },
  { to: '/#about', label: 'О курорте' },
  { to: '/#contacts', label: 'Контакты' },
];

export function Header() {
  const { user, initializing } = useAuth();
  const [open, setOpen] = useState(false);
  const location = useLocation();
  useEffect(() => setOpen(false), [location.pathname, location.hash]);

  return (
    <header className="header container">
      <div className="header__bar">
        <Logo />
        <nav className={`header__nav ${open ? 'header__nav--open' : ''}`} aria-label="Основная навигация">
          {NAV.map((n) =>
            n.to.includes('#') ? (
              <Link key={n.to} to={n.to} className="header__link">{n.label}</Link>
            ) : (
              <NavLink key={n.to} to={n.to} className={({ isActive }) => `header__link ${isActive ? 'header__link--active' : ''}`}>
                {n.label}
              </NavLink>
            ),
          )}
          <div className="header__nav-actions">
            <ButtonLink to={cabinetPath(user)} variant="light">Личный кабинет</ButtonLink>
            {!user && <ButtonLink to="/login">Войти</ButtonLink>}
          </div>
        </nav>
        <div className="header__actions">
          <ButtonLink to={cabinetPath(user)} variant="light">{user ? user.name.split(' ')[0] : 'Личный кабинет'}</ButtonLink>
          {!user && !initializing && <ButtonLink to="/login">Войти</ButtonLink>}
        </div>
        <button className="burger" aria-label="Меню" aria-expanded={open} onClick={() => setOpen((v) => !v)}>
          <span /><span /><span />
        </button>
      </div>
    </header>
  );
}

export function Footer() {
  return (
    <footer className="container">
      <div className="footer">
        <div className="footer__grid">
          <div className="footer__contacts">
            <div className="footer__label">Email</div>
            <a className="footer__email" href={`mailto:${CONTACTS.email}`}>{CONTACTS.email}</a>
            <div className="footer__label">Адрес</div>
            <div className="footer__address">{CONTACTS.address}<br />{CONTACTS.addressLine2}</div>
            <div className="footer__label">Соцсети</div>
            <div className="footer__socials">
              {CONTACTS.socials.map((s) => (
                <a key={s.label} href={s.href} className="social" target="_blank" rel="noreferrer">{s.label}</a>
              ))}
            </div>
          </div>
          <div className="footer__col">
            <div className="footer__heading">Занятия</div>
            <Link to="/lessons?category=ski">Горные лыжи</Link>
            <Link to="/lessons?category=snowboard">Сноуборд</Link>
            <Link to="/lessons?category=individual">Индивидуальное обучение</Link>
          </div>
          <div className="footer__col">
            <div className="footer__heading">Клиентам</div>
            <Link to="/rental">Прокат оборудования</Link>
            <Link to="/#instructors">Инструкторы</Link>
            <Link to="/account">Личный кабинет</Link>
          </div>
          <div className="footer__cta">
            <div className="footer__cta-title">Готовы<br />катать?</div>
            <ButtonLink to="/lessons" variant="light">Выбрать занятие</ButtonLink>
          </div>
        </div>
        <div className="footer__wordmark" aria-hidden>Montis</div>
        <div className="footer__bottom">
          <span>© {new Date().getFullYear()} Montis · Учебный проект</span>
          <span>Разработка: [имена команды]</span>
        </div>
      </div>
    </footer>
  );
}

/** Прокрутка к якорю (#about) и наверх при смене страницы */
function ScrollManager() {
  const { pathname, hash } = useLocation();
  useEffect(() => {
    if (hash) {
      const el = document.getElementById(hash.slice(1));
      if (el) { el.scrollIntoView({ behavior: 'smooth' }); return; }
    }
    window.scrollTo(0, 0);
  }, [pathname, hash]);
  return null;
}

export function PublicLayout() {
  return (
    <div className="page">
      <ScrollManager />
      <Header />
      <main className="page__main"><Outlet /></main>
      <Footer />
    </div>
  );
}
