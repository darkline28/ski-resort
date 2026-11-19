<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { instructorApi } from '@/api';
import type { Booking } from '@/types';

const bookings = ref<Booking[]>([]);
const error = ref('');
const loading = ref(true);

const now = new Date();

const today = [
  now.getFullYear(),
  String(now.getMonth() + 1).padStart(2, '0'),
  String(now.getDate()).padStart(2, '0'),
].join('-');

async function load() {
  loading.value = true;
  try {
    bookings.value = await instructorApi.bookings(today);
  } catch (e) {
    error.value = 'Не удалось загрузить';
  } finally {
    loading.value = false;
  }
}

async function complete(id: number) {
  try {
    await instructorApi.complete(id);
    await load();
  } catch (e) {
    error.value = 'Не удалось загрузить';
  }
}

onMounted(load);
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">Кабинет инструктора</h1>
    <p class="lead">Расписание занятий и отметка выполнения</p>

    <div class="card panel">
      <h2 class="h3">Занятия на сегодня</h2>

      <p v-if="loading">Загрузка...</p>
      <p v-else-if="error" class="state-error">{{ error }}</p>
      <p v-else-if="!bookings.length">Занятий пока нет.</p>

      <div v-for="b in bookings" :key="b.id" class="booking-line">
        <div>
          <b>{{ b.title }}</b>
          <p>
            {{ b.clientName }} · {{ new Date(b.start).toLocaleTimeString('ru-RU') }} ·
            {{ b.status }}
          </p>
        </div>

        <button
          v-if="b.status === 'confirmed'"
          class="btn btn--primary btn--sm"
          @click="complete(b.id)"
        >
          Завершить
        </button>
      </div>
    </div>
  </div>
</template>
