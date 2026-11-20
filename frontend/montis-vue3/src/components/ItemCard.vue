
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
    active?: boolean;
    category?: string;
  };
  kind: string;
}>();

const fallbackPhoto: Record<string, string> = {
  lessons: '/images/lesson-individual.jpg',
  equipment: '/images/eq-ski.jpg',
  instructors: '/images/instructor-ivan.jpg',
};

const isEquipmentSet = computed(
  () => props.kind === 'equipment' && props.item.category === 'kit'
);

const photo = computed(() => {
  return (
    props.item.imageUrl ||
    props.item.photoUrl ||
    fallbackPhoto[props.kind] ||
    '/images/eq-ski.jpg'
  );
});

const price = computed(
  () => props.item.price ?? props.item.pricePerHour
);

const formattedPrice = computed(() => {
  if (price.value === undefined || price.value === null) return '';

  return Number(price.value).toLocaleString('ru-RU');
});

const isUnavailable = computed(() => {
  return props.item.isActive === false || props.item.active === false;
});

const bookingQuery = computed(() => ({
  type: props.kind === 'equipment' ? 'rental' : 'lesson',
  id: String(props.item.id),
}));
</script>

<template>
  <article class="item-card">
    <img
      class="item-card__photo photo"
      :src="photo"
      :alt="item.title || item.name || 'Montis'"
    />

    <h3 class="item-card__title">
      {{ item.title || item.name }}
    </h3>

    <p class="item-card__text">
      {{ item.description || item.bio || item.specialization }}
    </p>

    <div v-if="item.durationMin" class="item-card__meta">
      {{ item.durationMin }} мин
    </div>

    <div
      v-if="kind === 'instructors' && item.experienceYears !== undefined"
      class="item-card__meta"
    >
      Опыт работы: {{ item.experienceYears }} лет
    </div>

    <div v-if="formattedPrice" class="item-card__price">
      от {{ formattedPrice }} ₽{{ kind === 'equipment' ? '/час' : '' }}
    </div>

    <button
      v-if="kind === 'instructors'"
      type="button"
      class="btn btn--primary"
      disabled
    >
      {{ isUnavailable ? 'Недоступен для записи' : 'Выбрать услугу' }}
    </button>

    <button
      v-else-if="isEquipmentSet"
      type="button"
      class="btn btn--primary"
      disabled
      title="Бронирование комплектов пока не подключено"
    >
      Скоро можно будет арендовать
    </button>

    <button
      v-else-if="isUnavailable"
      type="button"
      class="btn btn--primary"
      disabled
    >
      Недоступно
    </button>

    <RouterLink
      v-else
      class="btn btn--primary"
      :to="{
        path: '/booking',
        query: bookingQuery,
      }"
    >
      {{ kind === 'equipment' ? 'Арендовать' : 'Забронировать' }}
    </RouterLink>
  </article>
</template>
