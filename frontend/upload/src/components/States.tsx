/** Состояния загрузки, ошибки и пустого списка (экран 19 макета) */
import type { ReactNode } from 'react';
import type { AsyncState } from '@/hooks/useAsync';
import { Button, ButtonLink } from './ui';

export function Loader({ text = 'Загрузка данных' }: { text?: string }) {
  return (
    <div className="state state--loading" role="status">
      <span className="spinner" aria-hidden />
      <span>{text}</span>
    </div>
  );
}

export function SkeletonGrid({ count = 3 }: { count?: number }) {
  return (
    <div className="grid grid--3" aria-busy="true">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="skeleton-card">
          <div className="skeleton skeleton--img" />
          <div className="skeleton skeleton--line" />
          <div className="skeleton skeleton--line skeleton--short" />
        </div>
      ))}
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message?: string; onRetry?: () => void }) {
  return (
    <div className="state state--error" role="alert">
      <div className="state__icon state__icon--error">!</div>
      <h3>Ошибка загрузки</h3>
      <p>{message || 'Не удалось загрузить данные. Проверь соединение и попробуй ещё раз.'}</p>
      {onRetry && <Button variant="light" onClick={onRetry}>Повторить</Button>}
    </div>
  );
}

export function EmptyState({
  title = 'Пустой список',
  text = 'Пока ничего нет.',
  action,
}: { title?: string; text?: string; action?: { to: string; label: string } }) {
  return (
    <div className="state state--empty">
      <div className="state__icon">○</div>
      <h3>{title}</h3>
      <p>{text}</p>
      {action && <ButtonLink to={action.to}>{action.label}</ButtonLink>}
    </div>
  );
}

/**
 * Обёртка над useAsync: показывает лоадер / ошибку / пустое состояние / контент.
 */
export function AsyncView<T>({
  state, children, loading, empty, isEmpty,
}: {
  state: AsyncState<T>;
  children: (data: T) => ReactNode;
  loading?: ReactNode;
  empty?: ReactNode;
  isEmpty?: (data: T) => boolean;
}) {
  if (state.loading && state.data === undefined) return <>{loading ?? <Loader />}</>;
  if (state.error) return <ErrorState message={state.error.message} onRetry={state.reload} />;
  if (state.data === undefined) return null;
  const emptyCheck = isEmpty ?? ((d: T) => Array.isArray(d) && d.length === 0);
  if (empty && emptyCheck(state.data)) return <>{empty}</>;
  // при повторной загрузке показываем старые данные полупрозрачными
  if (state.loading) return <div className="refreshing" aria-busy="true">{children(state.data)}</div>;
  return <>{children(state.data)}</>;
}
