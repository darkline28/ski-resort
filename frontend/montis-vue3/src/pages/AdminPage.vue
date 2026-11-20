
<script setup lang="ts">
import { ref, watch } from 'vue';
import { catalogApi } from '@/api';
import type { Lesson, Equipment } from '@/types';

type Tab = 'bookings' | 'lessons' | 'equipment' | 'users';

const tab = ref<Tab>('lessons');

const lessons = ref<Lesson[]>([]);
const equipment = ref<Equipment[]>([]);

const error = ref('');
const loading = ref(false);

const tabs: { id: Tab; title: string }[] = [
  { id: 'bookings', title: 'Бронирования' },
  { id: 'lessons', title: 'Услуги' },
  { id: 'equipment', title: 'Оборудование' },
  { id: 'users', title: 'Пользователи' },
];

const unavailableTabs: Tab[] = ['bookings', 'users'];

function formatPrice(value: number): string {
  return Number(value).toLocaleString('ru-RU');
}

async function load() {
  error.value = '';

  if (unavailableTabs.includes(tab.value)) {
    loading.value = false;
    return;
  }

  loading.value = true;

  try {
    if (tab.value === 'lessons') {
      lessons.value = await catalogApi.lessons();
    }

    if (tab.value === 'equipment') {
      equipment.value = await catalogApi.equipment();
    }
  } catch {
    error.value = 'Не удалось загрузить';
  } finally {
    loading.value = false;
  }
}

watch(tab, load, { immediate: true });
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">Администрирование</h1>

    <p class="lead">
      Просмотр данных курорта и управление системой.
    </p>

    <div class="filters">
      <button
        v-for="t in tabs"
        :key="t.id"
        type="button"
        class="filter-pill"
        :class="{ 'filter-pill--active': tab === t.id }"
        @click="tab = t.id"
      >
        {{ t.title }}
      </button>
    </div>

    <div class="card panel">
      <template v-if="unavailableTabs.includes(tab)">
        <h2 class="h3">
          {{ tab === 'bookings' ? 'Бронирования' : 'Пользователи' }}
        </h2>

        <p class="muted">
          Этот раздел будет доступен после подключения
          соответствующих административных маршрутов API.
        </p>
      </template>

      <template v-else>
        <p v-if="loading" class="muted">
          Загрузка...
        </p>

        <template v-else-if="error">
          <p class="state-error">{{ error }}</p>

          <button
            type="button"
            class="btn btn--outline btn--sm"
            @click="load"
          >
            Повторить
          </button>
        </template>

        <div v-else class="table-wrap">
          <table class="table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Название</th>
                <th>Информация</th>
                <th>Статус</th>
              </tr>
            </thead>

            <tbody v-if="tab === 'lessons'">
              <tr v-for="lesson in lessons" :key="lesson.id">
                <td>{{ lesson.id }}</td>

                <td>{{ lesson.title }}</td>

                <td>
                  {{ formatPrice(lesson.price) }} ₽
                  · {{ lesson.durationMin }} мин
                </td>

                <td>
                  {{ lesson.active ? 'Активна' : 'Неактивна' }}
                </td>
              </tr>

              <tr v-if="!lessons.length">
                <td colspan="4">
                  Услуг пока нет.
                </td>
              </tr>
            </tbody>

            <tbody v-if="tab === 'equipment'">
              <tr v-for="item in equipment" :key="item.id">
                <td>{{ item.id }}</td>

                <td>{{ item.title }}</td>

                <td>
                  {{ formatPrice(item.pricePerHour) }} ₽/час
                  · {{ item.stock }} шт.
                </td>

                <td>
                  {{ item.category }}
                </td>
              </tr>

              <tr v-if="!equipment.length">
                <td colspan="4">
                  Оборудования пока нет.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <p class="muted small">
      Просмотр услуг и оборудования использует доступные маршруты API.
      Создание, редактирование и удаление будут подключены
      после реализации административных методов backend.
    </p>
  </div>
</template>
