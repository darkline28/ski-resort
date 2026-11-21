<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { catalogApi } from '@/api';
import { useRoute } from 'vue-router';
import ItemCard from '@/components/ItemCard.vue';
import type { Equipment, EquipmentCategory, Lesson } from '@/types';

const props = defineProps<{
  kind: 'lessons' | 'equipment' | 'instructors';
}>();
const route = useRoute();

type CatalogItem = Lesson | Equipment | {
  id: number;
  title?: string;
  name?: string;
  description?: string;
  bio?: string | null;
  imageUrl?: string | null;
  photoUrl?: string | null;
  price?: number;
  pricePerHour?: number;
  durationMin?: number;
  experienceYears?: number;
  active?: boolean;
  isActive?: boolean;
};

type CategoryOption = {
  code: string;
  name: string;
};

type BackendCategory = {
  id: number;
  name: string;
  code: string;
};

type BackendEquipmentSet = {
  id: number;
  name: string;
  description?: string;
  sport_type: 'ski' | 'snowboard';
  age_group: 'adult' | 'child';
  price_per_hour: number;
  is_active: boolean;
};

const items = ref<CatalogItem[]>([]);
const loading = ref(true);
const error = ref('');
const category = ref('all');

const equipmentCategories = ref<CategoryOption[]>([
  { code: 'all', name: 'Всё' },
]);

const titles = {
  lessons: [
    'Занятия на склоне',
    'Выбирай программу для своего уровня катания',
  ],
  equipment: [
    'Подготовься к покорению склонов',
    'Выбирай необходимое снаряжение для комфортного и безопасного катания. Всё оборудование доступно для аренды на удобный период',
  ],
  instructors: [
    'Наши инструкторы',
    'Опытные специалисты помогут освоить склоны',
  ],
};

const lessonCategories: CategoryOption[] = [
  { code: 'all', name: 'Все' },
];

const instructorCategories: CategoryOption[] = [
  { code: 'all', name: 'Все' },
];

const visibleCategories = computed(() => {
  if (props.kind === 'equipment') return equipmentCategories.value;
  if (props.kind === 'lessons') return lessonCategories;
  return instructorCategories;
});

async function loadEquipmentCategories() {
  try {
    const response = await catalogApi.equipmentCategories();

    equipmentCategories.value = [
      { code: 'all', name: 'Всё' },
      ...response.map((c: BackendCategory) => ({
        code: c.code,
        name: c.name,
      })),
      { code: 'kit', name: 'Комплекты' },
    ];
  } catch {
    equipmentCategories.value = [
      { code: 'all', name: 'Всё' },
      { code: 'ski', name: 'Лыжи' },
      { code: 'snowboard', name: 'Сноуборды' },
      { code: 'jacket', name: 'Куртки' },
      { code: 'pants', name: 'Штаны' },
      { code: 'boots', name: 'Обувь' },
      { code: 'helmet', name: 'Шлемы' },
      { code: 'goggles', name: 'Очки' },
      { code: 'kit', name: 'Комплекты' },
    ];
  }
}

function mapSet(set: BackendEquipmentSet): Equipment {
  return {
    id: set.id,
    title: set.name,
    description: set.description ?? '',
    category: 'kit',
    pricePerHour: set.price_per_hour,
    stock: 0,
    imageUrl: null,
    active: set.is_active,
  };
}

let loadVersion=0;
async function load() {
  const version=++loadVersion;
  loading.value = true;
  error.value = '';
  items.value = [];

  try {
    if (props.kind === 'lessons') {
      const id = Number(route.query.instructor);
      const result = (Number.isInteger(id) && id > 0 ? await catalogApi.instructorServices(id) : await catalogApi.lessons()).filter(x=>x.active);
      if(version===loadVersion) items.value=result;
    } else if (props.kind === 'instructors') {
      const result = (await catalogApi.instructors()).sort(
        (a, b) => Number(b.isActive) - Number(a.isActive)
      );
      if(version===loadVersion) items.value=result;
    } else if (category.value === 'kit') {
      const sets = await catalogApi.equipmentSets();
      if(version===loadVersion) items.value = sets
        .filter((set: BackendEquipmentSet) => set.is_active)
        .map(mapSet);
    } else {
      const result = await catalogApi.equipment(
        category.value === 'all'
          ? undefined
          : (category.value as EquipmentCategory)
      );
      if(version===loadVersion) items.value=result;
    }
  } catch {
    if(version===loadVersion) error.value = 'Не удалось загрузить';
  } finally {
    if(version===loadVersion) loading.value = false;
  }
}

watch(
  () => props.kind,
  async () => {
    category.value = 'all';

    if (props.kind === 'equipment') {
      await loadEquipmentCategories();
    }

    await load();
  },
  { immediate: true }
);

watch(category, load);
watch(()=>route.query.instructor,load);
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">{{ titles[kind][0] }}</h1>
    <p class="lead">{{ titles[kind][1] }}</p>

    <div class="filters">
      <button
        v-for="c in visibleCategories"
        :key="c.code"
        type="button"
        class="filter-pill"
        :class="{ 'filter-pill--active': category === c.code }"
        @click="category = c.code"
      >
        {{ c.name }}
      </button>
    </div>

    <div v-if="loading" class="card panel">
      Загружаем данные...
    </div>

    <div v-else-if="error" class="card panel empty-state">
      <div class="empty-state__symbol">↻</div>
      <p class="empty-state__title">{{ error }}</p>
      <button
        type="button"
        class="btn btn--outline btn--sm"
        @click="load"
      >
        Повторить
      </button>
    </div>

    <div v-else-if="!items.length" class="card panel">
      Пока нет доступных предложений.
    </div>

    <div v-else class="grid grid--3">
      <ItemCard
        v-for="item in items"
        :key="`${category}-${item.id}`"
        :item="item"
        :kind="kind"
      />
    </div>
  </div>
</template>
