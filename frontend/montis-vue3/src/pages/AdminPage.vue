
<script setup lang="ts">
import { ref, watch } from 'vue';
import { catalogApi, adminApi } from '@/api';
import AdminEntities from '@/components/AdminEntities.vue';
import AdminInstructors from '@/components/AdminInstructors.vue';
import type { Lesson, Equipment, Booking, BookingStatus, BookingType } from '@/types';

type Tab = 'bookings' | 'lessons' | 'equipment' | 'manage' | 'instructors';

const tab = ref<Tab>('lessons');

const lessons = ref<Lesson[]>([]);
const equipment = ref<Equipment[]>([]);

const error = ref('');
const loading = ref(false);

const tabs: { id: Tab; title: string }[] = [
  { id: 'bookings', title: 'Бронирования' },
  { id: 'lessons', title: 'Услуги' },
  { id: 'equipment', title: 'Оборудование' },
  { id: 'instructors', title: 'Инструкторы' },
  { id: 'manage', title: 'Управление справочниками' },
];

const bookings = ref<Booking[]>([]);
const bookingType = ref<BookingType>('lesson');
const updating = ref(false);
function transitions(b: Booking): BookingStatus[] {
  if (b.type === 'lesson') return b.status === 'confirmed' ? (Date.parse(b.start)>Date.now() ? ['cancelled'] : ['completed','no_show']) : [];
  if (b.status === 'pending') return ['confirmed','cancelled'];
  if (b.status === 'confirmed') return Date.parse(b.start)>Date.now() ? ['issued','cancelled'] : ['issued'];
  return b.status === 'issued' ? ['returned'] : [];
}
const labels: Record<BookingStatus,string> = {pending:'Ожидает',confirmed:'Подтверждено',cancelled:'Отменено',completed:'Завершено',no_show:'Неявка',issued:'Выдано',returned:'Возвращено'};
async function update(b: Booking, status: BookingStatus) {
  if (updating.value || !transitions(b).includes(status) || !confirm('Изменить статус бронирования?')) return;
  updating.value=true;
  try { await adminApi.setBookingStatus(b.id,b.type,status); await load(); }
  catch(e) { await load(); error.value=e instanceof Error?e.message:'Не удалось изменить статус'; }
  finally {updating.value=false;}
}

function formatPrice(value: number): string {
  return Number(value).toLocaleString('ru-RU');
}

async function load() {
  error.value = '';

  loading.value = true;

  try {
    if (tab.value === 'bookings') bookings.value = await adminApi.bookings(bookingType.value);
    if (tab.value === 'lessons') {
      lessons.value = await adminApi.services();
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

watch([tab,bookingType], load, { immediate: true });
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
      <AdminEntities v-if="tab === 'manage'" />
      <AdminInstructors v-else-if="tab === 'instructors'" />
      <template v-else-if="tab === 'bookings'">
        <label>Тип записи <select v-model="bookingType" class="field"><option value="lesson">Занятия</option><option value="rental">Аренда</option></select></label>
        <p v-if="loading">Загрузка...</p>
        <p v-else-if="error" class="state-error">{{error}} <button class="btn btn--outline" @click="load">Повторить</button></p>
        <p v-else-if="!bookings.length">Записей нет.</p>
        <div v-for="b in (!loading && !error ? bookings : [])" :key="b.id" class="booking-line">
          <div><b>{{b.title}}</b><p>{{b.clientName}} · {{new Date(b.start).toLocaleString('ru-RU',{timeZone:'Europe/Moscow'})}} · {{labels[b.status]}} · {{b.total}} ₽</p></div>
          <div><button v-for="s in transitions(b)" :key="s" class="btn btn--outline btn--sm" :disabled="updating" @click="update(b,s)">{{labels[s]}}</button></div>
        </div>
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

  </div>
</template>
