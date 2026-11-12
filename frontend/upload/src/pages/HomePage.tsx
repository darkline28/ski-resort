import { useState } from 'react';
import { catalogApi, feedbackApi } from '@/api';
import { useAsync } from '@/hooks/useAsync';
import { isEmail, isPhone, useForm } from '@/hooks/useForm';
import { useToast } from '@/context/ToastContext';
import { AsyncView, SkeletonGrid } from '@/components/States';
import { EquipmentCard, InstructorCard, LessonCard } from '@/components/Cards';
import { Alert, Button, ButtonLink, Field, Photo, Tag, TextArea } from '@/components/ui';
import { CONTACTS } from '@/config';

const ABOUT = [
  { title: 'Трассы', text: 'Подготовленные трассы разного уровня сложности', photo: 'Фото: трассы курорта', caption: 'Подготовленные трассы', long: 'Здесь тебя ждут трассы разного уровня сложности: от пологих учебных склонов до чёрных спусков.' },
  { title: 'Подъёмники', text: 'Современные подъёмники для быстрого доступа к склонам', photo: 'Фото: подъёмники', caption: 'Современные подъёмники', long: 'Кресельные и гондольные подъёмники доставят на вершину за считанные минуты.' },
  { title: 'Комфорт', text: 'Всё необходимое для комфортного отдыха', photo: 'Фото: инфраструктура', caption: 'Комфортный отдых', long: 'Кафе, тёплые раздевалки, камеры хранения и прокат — всё рядом со склоном.' },
];

export default function HomePage() {
  const lessons = useAsync(() => catalogApi.lessons(), []);
  const equipment = useAsync(() => catalogApi.equipment(), []);
  const instructors = useAsync(() => catalogApi.instructors(), []);
  const [aboutIdx, setAboutIdx] = useState(0);

  return (
    <>
      {/* Баннер */}
      <section className="container">
        <div className="hero">
          <svg className="hero__mountains" viewBox="0 0 1000 300" preserveAspectRatio="none" aria-hidden>
            <path d="M0 220 L180 120 L380 230 L520 80 L700 210 L780 170 L1000 260 V300 H0Z" fill="currentColor" opacity=".25" />
            <path d="M0 260 L250 180 L450 250 L650 150 L820 230 L1000 200 V300 H0Z" fill="currentColor" opacity=".35" />
          </svg>
          <div className="hero__content">
            <Tag dark>Горнолыжный курорт</Tag>
            <h1 className="hero__title">Открой для себя мир горных приключений!</h1>
            <p className="hero__text">
              Проведи незабываемое время на склонах, освой новые навыки и насладись красотой гор.
              Всё необходимое для идеального отдыха уже здесь.
            </p>
            <ButtonLink to="/lessons" variant="light">Выбрать занятие</ButtonLink>
          </div>
        </div>
      </section>

      {/* Популярные занятия */}
      <section className="container section" id="lessons">
        <div className="popular">
          <div className="popular__intro">
            <Tag dark>Популярные занятия</Tag>
            <h2 className="popular__title">Найди своё идеальное занятие</h2>
            <p>От первых шагов на лыжах до уверенного спуска по сложным трассам — у нас найдётся занятие для каждого.</p>
          </div>
          <AsyncView state={lessons} loading={<SkeletonGrid count={2} />}>
            {(list) => (
              <>
                {list.slice(0, 3).map((l) => <LessonCard key={l.id} lesson={l} />)}
                <div className="popular__more">
                  <h3>Больше занятий в каталоге</h3>
                  <p>Выбирай направление и уровень подготовки.</p>
                  <ButtonLink to="/lessons" variant="light" block>Все занятия</ButtonLink>
                </div>
              </>
            )}
          </AsyncView>
        </div>
      </section>

      {/* Прокат */}
      <section className="container section" id="rental">
        <SectionHead tag="Прокат оборудования" title="Всё для комфортного катания"
          text="Не нужно покупать собственное снаряжение. Выбирай подходящее оборудование и отправляйся покорять склоны."
          action={<ButtonLink to="/rental" variant="outline">Перейти к прокату</ButtonLink>} />
        <AsyncView state={equipment} loading={<SkeletonGrid />}>
          {(list) => <div className="grid grid--3">{list.slice(0, 3).map((e) => <EquipmentCard key={e.id} item={e} />)}</div>}
        </AsyncView>
      </section>

      {/* Инструкторы */}
      <section className="container section" id="instructors">
        <SectionHead tag="Инструкторы" title="Учись у профессионалов"
          text="Наши инструкторы помогут освоить технику катания, улучшить навыки и почувствовать уверенность на склоне."
          action={<ButtonLink to="/book/lesson" variant="outline">Записаться на занятие</ButtonLink>} />
        <AsyncView state={instructors} loading={<SkeletonGrid />}>
          {(list) => <div className="grid grid--3">{list.map((i) => <InstructorCard key={i.id} instructor={i} />)}</div>}
        </AsyncView>
      </section>

      {/* О курорте */}
      <section className="container section" id="about">
        <div className="about">
          <div className="about__left">
            <Tag dark>О курорте</Tag>
            <h2 className="about__title">Твой отдых начинается в горах</h2>
            <p className="about__lead">Наш горнолыжный курорт — место, где встречаются активный отдых, живописные пейзажи и яркие впечатления.</p>
            <ol className="about__list">
              {ABOUT.map((a, i) => (
                <li key={a.title}>
                  <button className={`about__item ${i === aboutIdx ? 'about__item--active' : ''}`} onClick={() => setAboutIdx(i)}>
                    <span className="about__num">{String(i + 1).padStart(2, '0')}</span>
                    <span>
                      <span className="about__item-title">{a.title}</span>
                      <span className="about__item-text">{a.text}</span>
                    </span>
                  </button>
                </li>
              ))}
            </ol>
          </div>
          <div className="about__right">
            <Photo alt={ABOUT[aboutIdx].caption} label={ABOUT[aboutIdx].photo} className="about__photo" />
            <div className="about__caption">{ABOUT[aboutIdx].caption}</div>
            <p className="about__caption-text">{ABOUT[aboutIdx].long}</p>
          </div>
        </div>
      </section>

      {/* Контакты */}
      <section className="container section" id="contacts">
        <SectionHead tag="Контакты" title="Мы всегда на связи"
          text="Остались вопросы? Свяжись с нами — поможем подобрать занятие, оборудование и ответим на все интересующие вопросы." />
        <div className="grid grid--2">
          <div className="card contacts-card">
            <div className="contacts-card__label">Адрес</div>
            <div className="contacts-card__value">{CONTACTS.address}</div>
            <div className="contacts-card__label">Телефон</div>
            <div className="contacts-card__value">{CONTACTS.phone}</div>
            <div className="contacts-card__label">Email</div>
            <a className="contacts-card__value" href={`mailto:${CONTACTS.email}`}>{CONTACTS.email}</a>
            <Photo alt="Карта курорта" label="Фото: карта или вид на курорт" className="contacts-card__map" />
          </div>
          <FeedbackForm />
        </div>
      </section>
    </>
  );
}

function SectionHead({ tag, title, text, action }: { tag: string; title: string; text: string; action?: React.ReactNode }) {
  return (
    <div className="section-head">
      <div>
        <Tag>{tag}</Tag>
        <h2 className="h2">{title}</h2>
        <p className="lead">{text}</p>
      </div>
      {action}
    </div>
  );
}

function FeedbackForm() {
  const toast = useToast();
  const [sent, setSent] = useState(false);
  const form = useForm({ name: '', contact: '', message: '' }, (v) => ({
    name: v.name.trim() ? undefined : 'Укажи имя',
    contact: isEmail(v.contact) || isPhone(v.contact) ? undefined : 'Укажи телефон или корректный email',
    message: v.message.trim().length >= 5 ? undefined : 'Напиши сообщение (минимум 5 символов)',
  }));

  return (
    <form className="card form" onSubmit={form.handleSubmit(async (v) => {
      await feedbackApi.send(v);
      setSent(true);
      form.reset();
      toast('Сообщение отправлено. Мы скоро ответим!');
    })} noValidate>
      <h3 className="h3">Форма обратной связи</h3>
      {sent && <Alert kind="success">Спасибо! Мы свяжемся с тобой в ближайшее время.</Alert>}
      {form.formError && <Alert>{form.formError}</Alert>}
      <Field name="name" placeholder="Имя" value={form.values.name} onChange={form.onChange} error={form.errors.name} />
      <Field name="contact" placeholder="Телефон или email" value={form.values.contact} onChange={form.onChange} error={form.errors.contact} />
      <TextArea name="message" placeholder="Сообщение" rows={4} value={form.values.message} onChange={form.onChange} error={form.errors.message} />
      <Button type="submit" block loading={form.submitting}>Отправить</Button>
    </form>
  );
}
