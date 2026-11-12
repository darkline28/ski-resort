import { Link } from 'react-router-dom';

export function Logo() {
  return (
    <Link to="/" className="logo" aria-label="Montis — на главную">
      <span className="logo__mark" aria-hidden>
        <svg viewBox="0 0 24 24" width="14" height="14"><path d="M12 4l9 15H3z" fill="currentColor" /></svg>
      </span>
      <span className="logo__text">Montis</span>
    </Link>
  );
}
