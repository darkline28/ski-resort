
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { catalogApi, bookingApi } from '@/api';
import { auth } from '@/stores/auth';
import type { Lesson, Equipment, Instructor, Slot } from '@/types';

const route = useRoute();
const router = useRouter();

const item = ref<Lesson | Equipment | null>(null);
const instructors = ref<Instructor[]>([]);
const slots = ref<Slot[]>([]);

const instructorId = ref(0);
const date = ref('');
const selectedSlot = ref('');
const rentalStart = ref('');
const rentalEnd = ref('');
const size = ref('');
const quantity = ref(1);

const error = ref('');
const message = ref('');
const loading = ref(false);
const submitting = ref(false);
const initializing = ref(false);

const rental = computed(() => route.query.type === 'rental');
const itemId = computed(() => Number(route.query.id));

const today = new Date().toLocaleDateString('en-CA');

const rentalSizes = computed(() => {
  if (!rental.value || !item.value) return [];
  return (item.value as Equipment).sizes ?? [];
});

const canSubmit = computed(() => {
  if (initializing.value || submitting.value || !item.value) return false;

  if (rental.value) {
    return (
      Boolean(rentalStart.value) &&
      Boolean(rentalEnd.value) &&
      quantity.value >= 1 &&
      (!rentalSizes.value.length || Boolean(size.value))
    );
  }

  return Boolean(instructorId.value && selectedSlot.value);
});

function toBackendDateTime(value: string): string {
  // datetime-local возвращает YYYY-MM-DDTHH:mm.
  // Добавляем секунды, если их нет.
  return value.length === 16 ? `${value}:00` : value;
}

async function init() {
  item.value = null;
  instructors.value = [];
  slots.value = [];
  instructorId.value = 0;
  selectedSlot.value = '';
  rentalStart.value = '';
  rentalEnd.value = '';
  size.value = '';
  quantity.value = 1;
  error.value = '';
  message.value = '';

  if (!Number.isInteger(itemId.value) || itemId.value <= 0) {
    error.value = 'Выберите услугу в каталоге';
    return;
  }

  initializing.value = true;

  try {
    item.value = rental.value
      ? await catalogApi.equipmentItem(itemId.value)
      : await catalogApi.lesson(itemId.value);

    if (rental.value) {
      size.value = rentalSizes.value[0] ?? '';
    } else {
      const allInstructors = (await catalogApi.instructors()).filter(
        (instructor) => instructor.isActive
      );

      const results = await Promise.all(
        allInstructors.map(async (instructor) => {
          try {
            const services = await catalogApi.instructorServices(
              instructor.id
            );

            return services.some(
              (service) => service.id === itemId.value && service.active
            )
              ? instructor
              : null;
          } catch {
            return null;
          }
        })
      );

      instructors.value = results.filter(
        (instructor): instructor is Instructor => instructor !== null
      );

      instructorId.value = instructors.value[0]?.id ?? 0;
    }
  } catch {
    error.value = 'Не удалось загрузить';
  } finally {
    initializing.value = false;
  }
}

async function loadSlots() {
  slots.value = [];
  selectedSlot.value = '';

  if (
    rental.value ||
    !date.value ||
    !instructorId.value ||
    !itemId.value
  ) {
    return;
  }

  loading.value = true;
  error.value = '';

  try {
    slots.value = await catalogApi.slots(
      instructorId.value,
      date.value,
      itemId.value
    );
  } catch {
    error.value = 'Не удалось загрузить свободное время';
  } finally {
    loading.value = false;
  }
}

watch(
  () => [route.query.id, route.query.type],
  init,
  { immediate: true }
);

watch([date, instructorId], loadSlots);

async function submit() {
  error.value = '';
  message.value = '';

  if (!item.value || !canSubmit.value) return;

  if (!auth.user) {
    await router.push({
      path: '/login',
      query: { next: route.fullPath },
    });
    return;
  }

  if (rental.value) {
    const startTime = new Date(rentalStart.value).getTime();
    const endTime = new Date(rentalEnd.value).getTime();

    if (
      !Number.isFinite(startTime) ||
      !Number.isFinite(endTime) ||
      endTime <= startTime
    ) {
      error.value = 'Конец аренды должен быть позже начала';
      return;
    }

    if (startTime <= Date.now()) {
      error.value = 'Выберите будущую дату начала аренды';
      return;
    }

    if (!Number.isInteger(quantity.value) || quantity.value < 1) {
      error.value = 'Укажите корректное количество';
      return;
    }
  }

  submitting.value = true;

  try {
    if (rental.value) {
      await bookingApi.createRental({
        equipmentId: itemId.value,
        size: size.value,
        quantity: quantity.value,
        start: toBackendDateTime(rentalStart.value),
        end: toBackendDateTime(rentalEnd.value),
      });
    } else {
      await bookingApi.createLesson({
        lessonId: itemId.value,
        instructorId: instructorId.value,
        start: selectedSlot.value,
      });
    }

    message.value = 'Бронирование создано';
    await router.push('/account');
  } catch {
    error.value = 'Не удалось создать бронирование';
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="container section section--top">
    <h1 class="h1">Оформление бронирования</h1>

    <p class="lead">
      Выберите удобное время и заполните данные для бронирования.
    </p>

    <div class="card panel booking-form">
      <h2 class="h3">{{ item?.title || 'Выбор услуги' }}</h2>

      <p v-if="initializing" class="muted">
        Загружаем данные...
      </p>

      <form
        v-else
        class="form-stack"
        @submit.prevent="submit"
      >
        <template v-if="!rental">
          <label>
            Инструктор
            <select
              v-model.number="instructorId"
              class="field"
              required
              :disabled="!instructors.length"
            >
              <option
                v-for="instructor in instructors"
                :key="instructor.id"
                :value="instructor.id"
              >
                {{ instructor.name }}
              </option>
            </select>
          </label>

          <p v-if="!instructors.length && item" class="muted">
            Для этого занятия пока нет доступных инструкторов.
          </p>

          <label>
            Дата занятия
            <input
              v-model="date"
              type="date"
              class="field"
              :min="today"
              required
            />
          </label>

          <p v-if="loading" class="muted">
            Загружаем свободное время...
          </p>

          <div v-if="slots.length" class="filters">
            <button
              v-for="slot in slots"
              :key="slot.start"
              type="button"
              class="filter-pill"
              :class="{
                'filter-pill--active': selectedSlot === slot.start
              }"
              @click="selectedSlot = slot.start"
            >
              {{
                new Date(slot.start).toLocaleTimeString('ru-RU', {
                  hour: '2-digit',
                  minute: '2-digit'
                })
              }}
            </button>
          </div>

          <p v-else-if="date && !loading && !error" class="muted">
            На выбранную дату свободного времени нет.
          </p>
        </template>

        <template v-else>
          <label v-if="rentalSizes.length">
            Размер
            <select
              v-model="size"
              class="field"
              required
            >
              <option
                v-for="option in rentalSizes"
                :key="option"
                :value="option"
              >
                {{ option }}
              </option>
            </select>
          </label>

          <label>
            Количество
            <input
              v-model.number="quantity"
              class="field"
              type="number"
              min="1"
              step="1"
              required
            />
          </label>

          <label>
            Начало аренды
            <input
              v-model="rentalStart"
              class="field"
              type="datetime-local"
              required
            />
          </label>

          <label>
            Конец аренды
            <input
              v-model="rentalEnd"
              class="field"
              type="datetime-local"
              :min="rentalStart || undefined"
              required
            />
          </label>

          <p v-if="!rentalSizes.length" class="muted">
            Размер для этого оборудования не указан.
            Доступность проверяется при оформлении.
          </p>
        </template>

        <p v-if="error" class="state-error">
          {{ error }}
        </p>

        <p v-if="message">{{ message }}</p>

        <button
          type="submit"
          class="btn btn--primary"
          :disabled="!canSubmit"
        >
          {{ submitting ? 'Оформляем...' : 'Забронировать' }}
        </button>
      </form>
    </div>
  </div>
</template>
