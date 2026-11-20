<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  item: {
    id: number | string;
    title?: string;
    name?: string;
    description?: string;
    bio?: string | null;
    imageUrl?: string | null;
    photoUrl?: string | null;
    price?: number;
    pricePerHour?: number;
    durationMin?: number;
    specialization?: string;
    experienceYears?: number;
    isActive?: boolean;
  };
  kind: string;
}>();

const fallbackPhoto: Record<string, string> = {
  lessons: '/images/lesson-individual.jpg',
  equipment: '/images/eq-ski.jpg',
  instructors: '/images/instructor-ivan.jpg',
};

const photo = computed(
  () => props.item.imageUrl || props.item.photoUrl || fallbackPhoto[props.kind],
);

const price = computed(() => props.item.price ?? props.item.pricePerHour);
</script>

<template>
  <article class="item-card">
    <img
      class="item-card__photo photo"
      :src="photo"
      :alt="item.title || item.name || 'Montis'"
    />

    <h3 class="item-card__title">{{ item.title || item.name }}</h3>

    <p class="item-card__text">
      {{ item.description || item.bio || item.specialization }}
    </p>

    <div v-if="item.durationMin" class="item-card__meta">
      {{ item.durationMin }} мин
    </div>

    <div v-if="kind === 'instructors' && item.experienceYears !== undefined" class="item-card__meta">
      Опыт работы: {{ item.experienceYears }} лет
    </div>

    <div v-if="price !== undefined" class="item-card__price">
      от {{ price.toLocaleString('ru-RU') }} ₽{{ kind === 'equipment' ? '/час' : '' }}
    </div>

    <button
  v-if="kind === 'instructors' && item.isActive === false"
  class="btn btn--primary"
  disabled
>
  Недоступен для записи
</button>

<RouterLink
  v-if="kind !== 'instructors'"
  class="btn btn--primary"
  :to="{
    path: '/booking',
    query: {
      type: kind === 'equipment' ? 'rental' : 'lesson',
      id: item.id,
    },
  }"
>
  {{ kind === 'equipment' ? 'Арендовать' : 'Забронировать' }}
</RouterLink>

<button
  v-else
  class="btn btn--primary"
  disabled
>
  {{ item.isActive === false ? 'Недоступен для записи' : 'Выбрать услугу' }}
</button>
    
  </article>
</template>
