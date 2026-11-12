import type { Equipment, Instructor, Lesson } from '@/types';
import { fmtPrice } from '@/utils/format';
import { ButtonLink, Photo } from './ui';

export function LessonCard({ lesson }: { lesson: Lesson }) {
  return (
    <article className="item-card">
      <Photo src={lesson.imageUrl} alt={lesson.title} label="Фото: занятие" className="item-card__photo" />
      <h3 className="item-card__title">{lesson.title}</h3>
      <p className="item-card__text">{lesson.description}</p>
      <div className="item-card__meta">{lesson.durationMin} мин</div>
      <div className="item-card__price">от {fmtPrice(lesson.price)}</div>
      <ButtonLink to={`/book/lesson/${lesson.id}`}>Забронировать</ButtonLink>
    </article>
  );
}

export function EquipmentCard({ item }: { item: Equipment }) {
  return (
    <article className="item-card">
      <Photo src={item.imageUrl} alt={item.title} label="Фото: снаряжение" className="item-card__photo item-card__photo--wide" />
      <h3 className="item-card__title">{item.title}</h3>
      <p className="item-card__text">{item.description}</p>
      <div className="item-card__price">от {item.pricePerHour.toLocaleString('ru-RU')} руб./час</div>
      <ButtonLink to={`/book/rental/${item.id}`}>Арендовать</ButtonLink>
    </article>
  );
}

export function InstructorCard({ instructor }: { instructor: Instructor }) {
  return (
    <article className="item-card">
      <Photo src={instructor.photoUrl} alt={instructor.name} label="Фото: инструктор" className="item-card__photo item-card__photo--tall" />
      <div className="instructor-card__row">
        <div>
          <h3 className="item-card__title">{instructor.name}</h3>
          <p className="item-card__text">{instructor.specialization}</p>
        </div>
        <div className="instructor-card__price">
          {fmtPrice(instructor.pricePerHour)}
          <small>за час</small>
        </div>
      </div>
      <ButtonLink to={`/book/lesson?instructor=${instructor.id}`}>Забронировать</ButtonLink>
    </article>
  );
}
