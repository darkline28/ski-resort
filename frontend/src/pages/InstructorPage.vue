
<script setup lang="ts">
import { ref, watch } from 'vue';
import { instructorApi } from '@/api';
import type { Booking } from '@/types';

const bookings = ref<Booking[]>([]);
const error = ref('');
const loading = ref(false);
const updating = ref(false);

function localDateString(date: Date): string {
  return new Intl.DateTimeFormat('sv-SE', {timeZone:'Europe/Moscow', year:'numeric', month:'2-digit', day:'2-digit'}).format(date);
}

const selectedDate = ref(localDateString(new Date()));

const statusLabels: Record<string, string> = {
  confirmed: 'Подтверждено',
  cancelled: 'Отменено',
  completed: 'Завершено',
  no_show: 'Неявка',
};

function statusLabel(status: string): string {
  return statusLabels[status] ?? status;
}

function formatTime(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return 'Время не указано';
  }

  return date.toLocaleTimeString('ru-RU', {
    timeZone: 'Europe/Moscow',
    hour: '2-digit',
    minute: '2-digit',
  });
}

async function load() {
  loading.value = true;
  error.value = '';
  bookings.value = [];

  try {
    bookings.value = await instructorApi.bookings(
      selectedDate.value
    );

    bookings.value.sort(
      (a, b) =>
        new Date(a.start).getTime() -
        new Date(b.start).getTime()
    );
  } catch {
    error.value = 'Не удалось загрузить';
  } finally {
    loading.value = false;
  }
}

function selectToday() {
  selectedDate.value = localDateString(new Date());
}
async function setStatus(booking: Booking, status: 'completed' | 'no_show') {
  if (updating.value || booking.status !== 'confirmed' || Date.parse(booking.start) > Date.now()) return;
  if (!confirm('Изменить статус занятия?')) return;
  updating.value = true;
  try {
    await instructorApi.setStatus(booking.id, status);
    await load();
  } catch (e) {
    await load();
    error.value = e instanceof Error ? e.message : 'Не удалось изменить статус';
  } finally { updating.value = false; }
}

watch(selectedDate, load, { immediate: true });
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">Кабинет инструктора</h1>

    <p class="lead">
      Просматривайте свои занятия и расписание на выбранную дату.
    </p>

    <div class="card panel">
      <h2 class="h3">Мои занятия</h2>

      <div class="form-stack">
        <label>
          Выберите дату
          <input
            v-model="selectedDate"
            type="date"
            class="field"
          />
        </label>

        <button
          type="button"
          class="btn btn--outline btn--sm"
          @click="selectToday"
        >
          Сегодня
        </button>
      </div>

      <p v-if="loading" class="muted">
        Загружаем занятия...
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

      <p v-else-if="!bookings.length" class="muted">
        На выбранную дату занятий нет.
      </p>

      <div
        v-for="booking in !loading && !error ? bookings : []"
        :key="booking.id"
        class="booking-line"
      >
        <div>
          <b>{{ booking.title }}</b>

          <p class="muted">
            Клиент: {{ booking.clientName || 'Не указан' }}
          </p>

          <p class="muted">
            {{ formatTime(booking.start) }} —
            {{ formatTime(booking.end) }}
            · {{ statusLabel(booking.status) }}
          </p>
        </div>

        <div v-if="booking.status === 'confirmed' && Date.parse(booking.start) <= Date.now()">
        <button
          type="button"
          class="btn btn--primary btn--sm"
          :disabled="updating"
          @click="setStatus(booking, 'completed')"
        >
          Завершить
        </button>
        <button type="button" class="btn btn--outline btn--sm" :disabled="updating" @click="setStatus(booking, 'no_show')">Неявка</button>
        </div>
      </div>
    </div>
  </div>
</template>
