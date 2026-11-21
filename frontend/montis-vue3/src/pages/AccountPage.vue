<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { bookingApi } from '@/api';
import { ApiError } from '@/api/client';
import { auth } from '@/stores/auth';
import { canCancelBooking } from '@/bookingRules';
import type { Booking } from '@/types';

type AccountBooking = Booking & {
  bookingType: 'lesson' | 'rental';
};

const bookings = ref<AccountBooking[]>([]);
const error = ref('');
const actionError = ref('');
const loading = ref(true);
const cancellingId = ref<string | null>(null);

const hasBookings = computed(() => bookings.value.length > 0);

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'Ожидает подтверждения',
    confirmed: 'Подтверждено',
    cancelled: 'Отменено',
    completed: 'Завершено',
    no_show: 'Неявка',
    issued: 'Выдано',
    returned: 'Возвращено',
  };

  return labels[status] ?? status;
}

function formatDate(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return 'Дата не указана';
  }

  return date.toLocaleString('ru-RU', {
    timeZone: 'Europe/Moscow',
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatPrice(value?: number): string {
  if (value == null) return '—';

  return `${value.toLocaleString('ru-RU')} ₽`;
}

function bookingKey(booking: AccountBooking): string {
  return `${booking.bookingType}-${booking.id}`;
}

function canCancel(booking: AccountBooking): boolean {
  return canCancelBooking(booking.bookingType,booking.status,booking.start);
}

async function load() {
  loading.value = true;
  error.value = '';

  try {
    const result = await bookingApi.my();
    error.value = result.errors.join('. ');
    bookings.value = result.bookings.map((booking) => ({
      ...booking,
      bookingType:
        booking.bookingType === 'rental' ||
        booking.type === 'rental'
          ? 'rental'
          : 'lesson',
    }));
  } catch {
    error.value = 'Не удалось загрузить';
    bookings.value = [];
  } finally {
    loading.value = false;
  }
}

async function cancel(booking: AccountBooking) {
  if (cancellingId.value || loading.value || !canCancel(booking)) return;
  if (!confirm('Отменить бронирование?')) return;

  actionError.value = '';
  cancellingId.value = bookingKey(booking);

  try {
    await bookingApi.cancel(booking.id, booking.bookingType);
    await load();
  } catch (e) {
    actionError.value = e instanceof ApiError && e.status === 409
      ? e.message
      : 'Не удалось отменить бронирование. Проверьте состояние записи перед повтором.';
    await load();
  } finally {
    cancellingId.value = null;
  }
}

onMounted(load);
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">Личный кабинет</h1>

    <p class="lead">
      Здравствуйте, {{ auth.user?.name }}! Здесь ваши бронирования.
    </p>

    <div class="card panel">
      <h2 class="h3">Мои записи</h2>

      <p v-if="loading">Загрузка...</p>
      <p v-if="actionError" class="state-error">{{ actionError }}</p>

      <template v-if="!loading && error && !hasBookings">
        <p class="state-error">{{ error }}</p>

        <button
          type="button"
          class="btn btn--outline btn--sm"
          @click="load"
        >
          Попробовать снова
        </button>
      </template>

      <p v-else-if="!loading && !hasBookings">
        Записей пока нет.
      </p>

      <template v-else-if="!loading">
        <p v-if="error" class="state-error">{{ error }} <button type="button" class="btn btn--outline btn--sm" @click="load">Повторить</button></p>

        <div
          v-for="booking in bookings"
          :key="bookingKey(booking)"
          class="booking-line"
        >
          <div>
            <b>{{ booking.title }}</b>

            <p class="muted">
              {{
                booking.bookingType === 'lesson'
                  ? 'Занятие'
                  : 'Аренда оборудования'
              }}
              · {{ formatDate(booking.start) }}
              — {{ formatDate(booking.end) }} (МСК)
              · {{ statusLabel(booking.status) }}
              · {{ formatPrice(booking.total) }}
            </p>
          </div>

          <button
            v-if="canCancel(booking)"
            type="button"
            class="btn btn--outline btn--sm"
            :disabled="cancellingId !== null"
            @click="cancel(booking)"
          >
            {{
              cancellingId === bookingKey(booking)
                ? 'Отменяем...'
                : 'Отменить'
            }}
          </button>
        </div>
      </template>
    </div>
  </div>
</template>
