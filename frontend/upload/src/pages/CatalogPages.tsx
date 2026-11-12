import { useSearchParams } from 'react-router-dom';
import { catalogApi } from '@/api';
import { useAsync } from '@/hooks/useAsync';
import { AsyncView, EmptyState, SkeletonGrid } from '@/components/States';
import { EquipmentCard, LessonCard } from '@/components/Cards';
import { Chips } from '@/components/ui';
import type { EquipmentCategory, LessonCategory } from '@/types';

const LESSON_FILTERS: { value: LessonCategory | 'all'; label: string }[] = [
  { value: 'all', label: 'Все' },
  { value: 'ski', label: 'Горные лыжи' },
  { value: 'snowboard', label: 'Сноуборд' },
  { value: 'individual', label: 'Индивидуальное обучение' },
  { value: 'group', label: 'Групповые' },
  { value: 'kids', label: 'Детские' },
  { value: 'freestyle', label: 'Фристайл' },
];

const EQUIPMENT_FILTERS: { value: EquipmentCategory | 'all'; label: string }[] = [
  { value: 'all', label: 'Всё' },
  { value: 'ski', label: 'Лыжи' },
  { value: 'snowboard', label: 'Сноуборды' },
  { value: 'boots', label: 'Ботинки' },
  { value: 'helmet', label: 'Шлемы' },
  { value: 'goggles', label: 'Маски' },
  { value: 'kit', label: 'Комплекты' },
];

/** Фильтр хранится в URL (?category=ski), чтобы ссылки из футера работали */
function useCategory<T extends string>() {
  const [params, setParams] = useSearchParams();
  const value = (params.get('category') as T | null) ?? 'all';
  const set = (v: T | 'all') => setParams(v === 'all' ? {} : { category: v }, { replace: true });
  return [value as T | 'all', set] as const;
}

export function LessonsPage() {
  const [category, setCategory] = useCategory<LessonCategory>();
  const state = useAsync(() => catalogApi.lessons(category === 'all' ? undefined : category), [category]);

  return (
    <section className="container section section--top">
      <h1 className="h1">Занятия на любой уровень подготовки</h1>
      <p className="lead">Выбирай направление, подходящее именно тебе. Наши инструкторы помогут сделать первые шаги или улучшить уже имеющиеся навыки.</p>
      <div className="filters"><Chips options={LESSON_FILTERS} value={category} onChange={setCategory} /></div>
      <AsyncView state={state} loading={<SkeletonGrid count={6} />}
        empty={<EmptyState title="Занятий не найдено" text="В этой категории пока нет занятий. Попробуй другую." />}>
        {(list) => <div className="grid grid--3">{list.map((l) => <LessonCard key={l.id} lesson={l} />)}</div>}
      </AsyncView>
    </section>
  );
}

export function RentalPage() {
  const [category, setCategory] = useCategory<EquipmentCategory>();
  const state = useAsync(() => catalogApi.equipment(category === 'all' ? undefined : category), [category]);

  return (
    <section className="container section section--top">
      <h1 className="h1">Подготовься к покорению склонов</h1>
      <p className="lead">Выбирай необходимое снаряжение для комфортного и безопасного катания. Всё оборудование доступно для аренды на удобный период.</p>
      <div className="filters"><Chips options={EQUIPMENT_FILTERS} value={category} onChange={setCategory} /></div>
      <AsyncView state={state} loading={<SkeletonGrid count={6} />}
        empty={<EmptyState title="Ничего не найдено" text="В этой категории пока нет снаряжения." />}>
        {(list) => <div className="grid grid--3">{list.map((e) => <EquipmentCard key={e.id} item={e} />)}</div>}
      </AsyncView>
    </section>
  );
}
