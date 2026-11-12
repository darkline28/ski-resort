/** Базовые UI-компоненты: кнопки, поля, чипы, бейджи, фото-заглушки, модалка */
import { useEffect, type ButtonHTMLAttributes, type InputHTMLAttributes, type ReactNode, type TextareaHTMLAttributes } from 'react';
import { Link, type LinkProps } from 'react-router-dom';
import type { BookingStatus } from '@/types';
import { STATUS_LABEL } from '@/utils/format';

type Variant = 'primary' | 'light' | 'outline' | 'ghost' | 'danger';

const cls = (...c: (string | false | null | undefined)[]) => c.filter(Boolean).join(' ');

export function Button({
  variant = 'primary', block, loading, children, className, ...rest
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; block?: boolean; loading?: boolean }) {
  return (
    <button
      className={cls('btn', `btn--${variant}`, block && 'btn--block', className)}
      disabled={loading || rest.disabled}
      {...rest}
    >
      {loading ? <span className="spinner spinner--sm" aria-hidden /> : null}
      {children}
    </button>
  );
}

export function ButtonLink({ variant = 'primary', block, className, ...rest }: LinkProps & { variant?: Variant; block?: boolean }) {
  return <Link className={cls('btn', `btn--${variant}`, block && 'btn--block', className)} {...rest} />;
}

export function Field({
  label, error, className, ...rest
}: InputHTMLAttributes<HTMLInputElement> & { label?: string; error?: string }) {
  const id = rest.id ?? rest.name;
  return (
    <label className={cls('field', error && 'field--error', className)} htmlFor={id}>
      {label && <span className="field__label">{label}</span>}
      <input id={id} className="input" aria-invalid={!!error} {...rest} placeholder={rest.placeholder ?? (label ? undefined : rest.name)} />
      {error && <span className="field__error">{error}</span>}
    </label>
  );
}

export function TextArea({
  label, error, ...rest
}: TextareaHTMLAttributes<HTMLTextAreaElement> & { label?: string; error?: string }) {
  const id = rest.id ?? rest.name;
  return (
    <label className={cls('field', error && 'field--error')} htmlFor={id}>
      {label && <span className="field__label">{label}</span>}
      <textarea id={id} className="input input--area" aria-invalid={!!error} {...rest} />
      {error && <span className="field__error">{error}</span>}
    </label>
  );
}

export function Chips<T extends string | number>({
  options, value, onChange, disabled,
}: {
  options: { value: T; label: ReactNode; disabled?: boolean }[];
  value: T | null | undefined;
  onChange: (v: T) => void;
  disabled?: boolean;
}) {
  return (
    <div className="chips" role="radiogroup">
      {options.map((o) => (
        <button
          key={String(o.value)}
          type="button"
          role="radio"
          aria-checked={o.value === value}
          disabled={disabled || o.disabled}
          className={cls('chip', o.value === value && 'chip--active')}
          onClick={() => onChange(o.value)}
        >
          {o.label}
        </button>
      ))}
    </div>
  );
}

export function StatusBadge({ status }: { status: BookingStatus }) {
  return <span className={`badge badge--${status}`}>{STATUS_LABEL[status]}</span>;
}

export function Tag({ children, dark }: { children: ReactNode; dark?: boolean }) {
  return <span className={cls('tag', dark && 'tag--dark')}>{children}</span>;
}

/** Фото. Если url нет — заглушка с горами, как в макете */
export function Photo({ src, alt, label, className }: { src?: string | null; alt: string; label?: string; className?: string }) {
  if (src) return <img className={cls('photo', className)} src={src} alt={alt} loading="lazy" />;
  return (
    <div className={cls('photo photo--placeholder', className)} role="img" aria-label={alt}>
      {label && <span className="photo__label">{label}</span>}
      <svg className="photo__mountains" viewBox="0 0 400 140" preserveAspectRatio="none" aria-hidden>
        <path d="M0 80 L70 40 L130 85 L210 10 L290 80 L350 45 L400 75 V140 H0Z" fill="currentColor" opacity=".45" />
        <path d="M0 110 L90 70 L170 105 L250 65 L330 110 L400 90 V140 H0Z" fill="currentColor" opacity=".8" />
      </svg>
    </div>
  );
}

export function Modal({ open, onClose, title, children }: { open: boolean; onClose: () => void; title: string; children: ReactNode }) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);
  if (!open) return null;
  return (
    <div className="modal" onMouseDown={(e) => { if (e.target === e.currentTarget) onClose(); }}>
      <div className="modal__box card" role="dialog" aria-modal="true" aria-label={title}>
        <div className="modal__head">
          <h3>{title}</h3>
          <button className="modal__close" onClick={onClose} aria-label="Закрыть">×</button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function Alert({ kind = 'error', children }: { kind?: 'error' | 'success' | 'info'; children: ReactNode }) {
  return <div className={`alert alert--${kind}`} role={kind === 'error' ? 'alert' : 'status'}>{children}</div>;
}
