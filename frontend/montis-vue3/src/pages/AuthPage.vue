<script setup lang="ts">
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { auth, signIn, signUp } from '@/stores/auth';

const props = defineProps<{
  mode: 'login' | 'register';
}>();

const router = useRouter();
const route = useRoute();

const name = ref('');
const phone = ref('');
const email = ref('');
const password = ref('');

const error = ref('');
const loading = ref(false);

async function submit() {
  if (loading.value) return;

  error.value = '';
  loading.value = true;

  try {
    if (props.mode === 'login') {
      await signIn(
        email.value.trim(),
        password.value
      );
    } else {
      await signUp(
        name.value.trim(),
        phone.value.trim(),
        email.value.trim(),
        password.value
      );
    }

    const next =
      typeof route.query.next === 'string' &&
      route.query.next.startsWith('/') &&
      !route.query.next.startsWith('//')
        ? route.query.next
        : auth.user?.role === 'admin'
          ? '/admin'
          : auth.user?.role === 'instructor'
            ? '/instructor'
            : '/account';

    await router.push(next);
  } catch (e) {
    error.value =
      e instanceof Error
        ? e.message
        : 'Не удалось выполнить запрос';
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="container section">
    <div class="card panel form-card">
      <h1 class="h2">
        {{ mode === 'login' ? 'Вход в аккаунт' : 'Регистрация' }}
      </h1>

      <form class="form-stack" @submit.prevent="submit">
        <label v-if="mode === 'register'">
          Имя
          <input
            v-model.trim="name"
            type="text"
            class="field"
            minlength="2"
            autocomplete="name"
            required
          />
        </label>

        <label v-if="mode === 'register'">
          Телефон
          <input
            v-model.trim="phone"
            type="tel"
            class="field"
            placeholder="+79991234567"
            autocomplete="tel"
            required
          />
        </label>

        <label>
          Email
          <input
            v-model.trim="email"
            type="email"
            class="field"
            autocomplete="email"
            required
          />
        </label>

        <label>
          Пароль
          <input
            v-model="password"
            type="password"
            class="field"
            minlength="6"
            :autocomplete="
              mode === 'login' ? 'current-password' : 'new-password'
            "
            required
          />
        </label>

        <p v-if="error" class="state-error">
          {{ error }}
        </p>

        <button
          type="submit"
          class="btn btn--primary"
          :disabled="loading"
        >
          {{
            loading
              ? 'Подождите...'
              : mode === 'login'
                ? 'Войти'
                : 'Зарегистрироваться'
          }}
        </button>
      </form>

      <RouterLink
        class="link"
        :to="mode === 'login' ? '/register' : '/login'"
      >
        {{
          mode === 'login'
            ? 'Создать аккаунт'
            : 'Уже есть аккаунт? Войти'
        }}
      </RouterLink>
    </div>
  </div>
</template>
