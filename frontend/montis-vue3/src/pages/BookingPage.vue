
<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { catalogApi, bookingApi } from '@/api';
import { ApiError } from '@/api/client';
import { auth } from '@/stores/auth';
import { validRentalPeriod, validQuantity } from '@/bookingRules';
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
const availability = ref<{ size: string; available_quantity: number }[]>([]);
const availabilityLoading = ref(false);
const availabilityChecked = ref(false);

const error = ref('');
const message = ref('');
const loading = ref(false);
const submitting = ref(false);
const initializing = ref(false);

const isSet = computed(() => route.query.type === 'set');
const rental = computed(() => route.query.type === 'rental' || isSet.value);
const setItems = ref<{category_id:number;category_name:string;quantity:number}[]>([]);
const setSizes = ref<Record<number,string>>({});
const setOptions = ref<Record<number,{size:string;available_quantity:number}[]>>({});
const itemId = computed(() => Number(route.query.id));

const today = new Intl.DateTimeFormat('sv-SE', { timeZone: 'Europe/Moscow', year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date());

const rentalSizes = computed(() => availability.value.filter((row) => row.available_quantity > 0));
const selectedAvailable = computed(() => rentalSizes.value.find((row) => row.size === size.value)?.available_quantity ?? 0);
const rentalDatesValid = computed(() => validRentalPeriod(rentalStart.value, rentalEnd.value));
function toBackendDateTime(value: string): string {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(value)) throw new Error('Некорректная дата');
  return `${value}:00+03:00`;
}
let availabilityRequest = 0;
let initRequest = 0;
async function checkAvailability() {
  const requestId = ++availabilityRequest;
  availability.value = [];
  size.value = '';
  setSizes.value = {};
  setOptions.value = {};
  availabilityChecked.value = false;
  availabilityLoading.value = false;
  if (!rental.value || !rentalDatesValid.value || !itemId.value) return;
  availabilityLoading.value = true;
  error.value = '';
  try {
    const rows = await catalogApi.equipmentAvailability(isSet.value ? undefined : itemId.value, toBackendDateTime(rentalStart.value), toBackendDateTime(rentalEnd.value));
    if (requestId !== availabilityRequest) return;
    if (isSet.value) {
      const [models, categories] = await Promise.all([catalogApi.equipment(), catalogApi.equipmentCategories()]);
      if (requestId !== availabilityRequest) return;
      for (const entry of setItems.value) {
        const category = categories.find(c=>c.id===entry.category_id);
        const ids = new Set(models.filter(m=>m.category===category?.code || m.category===category?.name).map(m=>m.id));
        const counts = new Map<string,number>();
        for (const row of rows.filter(r=>ids.has(r.equipment_id))) counts.set(row.size,(counts.get(row.size)??0)+row.available_quantity);
        setOptions.value[entry.category_id] = [...counts].filter(([,n])=>n>=entry.quantity).map(([size,n])=>({size,available_quantity:Math.floor(n/entry.quantity)}));
      }
      availabilityChecked.value = true;
      return;
    }
    availability.value = rows.filter((row) => row.equipment_id === itemId.value);
    size.value = availability.value.find((row) => row.available_quantity > 0)?.size ?? '';
    availabilityChecked.value = true;
  } catch (e) {
    if (requestId !== availabilityRequest) return;
    error.value = e instanceof ApiError && e.status === 404
      ? 'Оборудование не найдено. Обновите каталог.'
      : 'Не удалось проверить наличие оборудования. Повторите попытку.';
  } finally {
    if (requestId === availabilityRequest) availabilityLoading.value = false;
  }
}
watch([rentalStart, rentalEnd, itemId], checkAvailability);

const canSubmit = computed(() => {
  if (initializing.value || submitting.value || loading.value || !item.value || !item.value.active) return false;

  if (rental.value) {
    if (isSet.value) return rentalDatesValid.value && availabilityChecked.value && !availabilityLoading.value && Number.isInteger(quantity.value) && quantity.value>0 && setItems.value.length>0 && setItems.value.every(e=>setOptions.value[e.category_id]?.some(o=>o.size===setSizes.value[e.category_id] && o.available_quantity>=quantity.value));
    return rentalDatesValid.value && availabilityChecked.value && !availabilityLoading.value &&
      Boolean(size.value) && validQuantity(quantity.value, selectedAvailable.value);
  }

  return Boolean(instructorId.value && selectedSlot.value && slots.value.some((slot) => slot.start === selectedSlot.value && Date.parse(slot.start) > Date.now()));
});

async function init() {
  const requestId = ++initRequest;
  ++slotRequest;
  ++availabilityRequest;
  loading.value = false;
  availabilityLoading.value = false;
  initializing.value = false;
  item.value = null;
  instructors.value = [];
  slots.value = [];
  instructorId.value = 0;
  selectedSlot.value = '';
  rentalStart.value = '';
  rentalEnd.value = '';
  size.value = '';
  quantity.value = 1;
  availability.value = [];
  availabilityChecked.value = false;
  error.value = '';
  message.value = '';

  if (!['lesson','rental','set'].includes(String(route.query.type)) || !Number.isInteger(itemId.value) || itemId.value <= 0) {
    error.value = 'Выберите услугу в каталоге';
    return;
  }

  initializing.value = true;

  try {
    if (isSet.value) {
      const s = await catalogApi.equipmentSet(itemId.value);
      if (requestId !== initRequest) return;
      setItems.value = s.items;
      item.value = {id:s.id,title:s.name,description:s.description??'',category:'kit',pricePerHour:s.price_per_hour,stock:0,active:s.is_active};
    } else {
      const result = rental.value ? await catalogApi.equipmentItem(itemId.value) : await catalogApi.lesson(itemId.value);
      if (requestId !== initRequest) return;
      item.value = result;
    }

    if (rental.value) {
      await checkAvailability();
    } else {
      const allInstructors = (await catalogApi.instructors()).filter(
        (instructor) => instructor.isActive
      );
      if (requestId !== initRequest) return;

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
      if (requestId !== initRequest) return;

      instructors.value = results.filter(
        (instructor): instructor is Instructor => instructor !== null
      );

      const preferred = Number(route.query.instructor);
      instructorId.value = instructors.value.find(i=>i.id===preferred)?.id ?? instructors.value[0]?.id ?? 0;
    }
  } catch {
    if (requestId === initRequest) error.value = 'Не удалось загрузить';
  } finally {
    if (requestId === initRequest) initializing.value = false;
  }
}

let slotRequest = 0;
async function loadSlots() {
  const requestId = ++slotRequest;
  loading.value = false;
  slots.value = [];
  selectedSlot.value = '';

  if (
    rental.value ||
    !date.value || date.value < today ||
    !instructorId.value ||
    !itemId.value
  ) {
    return;
  }

  loading.value = true;
  error.value = '';

  try {
    const result = await catalogApi.slots(
      instructorId.value,
      date.value,
      itemId.value
    );
    if (requestId === slotRequest) slots.value = result.filter(slot => Date.parse(slot.start) > Date.now());
  } catch {
    if (requestId === slotRequest) error.value = 'Не удалось загрузить свободное время';
  } finally {
    if (requestId === slotRequest) loading.value = false;
  }
}

watch(
  () => [route.query.id, route.query.type, route.query.instructor],
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

  if (auth.user.role !== 'client') {
    error.value = 'Бронирование доступно только клиентам.';
    return;
  }

  if (rental.value) {
    const startTime = Date.parse(toBackendDateTime(rentalStart.value));
    const endTime = Date.parse(toBackendDateTime(rentalEnd.value));

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

    if (!validRentalPeriod(rentalStart.value, rentalEnd.value)) {
      error.value = 'Аренда должна начинаться в будущем и длиться целое число часов (не менее одного).';
      return;
    }
    if (!Number.isInteger(quantity.value) || quantity.value < 1 || (!isSet.value && quantity.value > selectedAvailable.value)) {
      error.value = 'Укажите корректное количество';
      return;
    }
  } else if (Date.parse(selectedSlot.value) <= Date.now()) {
    error.value = 'Время занятия уже прошло. Выберите другой слот.';
    await loadSlots();
    return;
  }

  submitting.value = true;

  try {
    if (isSet.value) {
      await bookingApi.createSet(itemId.value,quantity.value,setItems.value.map(e=>({category_id:e.category_id,size:setSizes.value[e.category_id]})),toBackendDateTime(rentalStart.value),toBackendDateTime(rentalEnd.value));
    } else if (rental.value) {
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
  } catch (e) {
    if (e instanceof ApiError && e.status === 409) {
      if (rental.value) await checkAvailability();
      else await loadSlots();
      error.value = e.message;
    } else if (e instanceof ApiError && e.status === 401) {
      error.value = 'Сессия истекла. Войдите заново и проверьте бронирование.';
    } else {
      error.value = e instanceof ApiError && e.status === 0
        ? 'Связь прервалась. Проверьте список бронирований перед повторной отправкой.'
        : e instanceof Error ? e.message : 'Не удалось создать бронирование';
    }
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
                  timeZone: 'Europe/Moscow',
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
          <template v-if="isSet && availabilityChecked">
            <label v-for="entry in setItems" :key="entry.category_id">
              {{ entry.category_name }} — размер
              <select v-model="setSizes[entry.category_id]" class="field" required>
                <option value="">Выберите размер</option>
                <option v-for="o in (setOptions[entry.category_id] || [])" :key="o.size" :value="o.size">{{ o.size }} (комплектов: {{ o.available_quantity }})</option>
              </select>
              <span v-if="!setOptions[entry.category_id]?.length" class="state-error">Нет в наличии</span>
            </label>
          </template>
          <label v-if="!isSet && rentalSizes.length">
            Размер
            <select
              v-model="size"
              class="field"
              required
            >
              <option
                v-for="option in rentalSizes"
                :key="option.size"
                :value="option.size"
              >
                {{ option.size }} (доступно: {{ option.available_quantity }})
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
              :max="selectedAvailable || undefined"
              step="1"
              required
            />
          </label>

          <p class="muted">Время указывается по Москве (МСК).</p>
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

          <p v-if="availabilityLoading" class="muted">Проверяем наличие...</p>
          <p v-else-if="!isSet && rentalDatesValid && availabilityChecked && !rentalSizes.length" class="muted">На выбранный период оборудование недоступно.</p>
          <button v-if="rentalDatesValid && !availabilityLoading" type="button" class="btn btn--outline btn--sm" @click="checkAvailability">Обновить наличие</button>
          <p v-if="rentalStart && rentalEnd && !rentalDatesValid" class="state-error">Выберите будущий период длительностью целое число часов.</p>
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
