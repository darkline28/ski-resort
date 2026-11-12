import { ButtonLink } from '@/components/ui';

export default function NotFoundPage() {
  return (
    <section className="container section section--top">
      <div className="card result">
        <div className="result__icon">404</div>
        <h1 className="h2">Страница не найдена</h1>
        <p className="lead">Похоже, эта трасса закрыта. Вернись на главную и выбери другой маршрут.</p>
        <ButtonLink to="/">На главную</ButtonLink>
      </div>
    </section>
  );
}
