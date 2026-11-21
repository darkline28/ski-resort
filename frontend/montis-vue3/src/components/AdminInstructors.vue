<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { adminApi } from '@/api';
import type { Instructor, Lesson } from '@/types';

const instructors=ref<Instructor[]>([]), clients=ref<Awaited<ReturnType<typeof adminApi.clients>>>([]), services=ref<Lesson[]>([]), assigned=ref<Lesson[]>([]);
const selected=ref(0), clientId=ref(0), specialization=ref(''), experience=ref<number|null>(null), description=ref(''), serviceId=ref(0);
const appointing=ref(false), busy=ref(false), loading=ref(false), error=ref(''), success=ref('');
const current=computed(()=>instructors.value.find(x=>x.id===selected.value));
const availableServices=computed(()=>services.value.filter(s=>!assigned.value.some(a=>a.id===s.id)));
async function load(){
  loading.value=true;error.value='';
  try{
    const [i,c,s]=await Promise.all([adminApi.instructors(),adminApi.clients(),adminApi.services()]);
    instructors.value=i;clients.value=c;services.value=s;
    if(!i.some(x=>x.id===selected.value))selected.value=i[0]?.id??0;
    await loadAssignments();
  }catch(e){error.value=e instanceof Error?e.message:'Не удалось загрузить данные';}
  finally{loading.value=false;}
}
let version=0;
async function loadAssignments(){
  const v=++version;assigned.value=[];serviceId.value=0;
  if(!selected.value)return;
  try{const rows=await adminApi.instructorServices(selected.value);if(v===version)assigned.value=rows;}
  catch(e){if(v===version)error.value=e instanceof Error?e.message:'Не удалось загрузить назначения';}
}
watch(selected,loadAssignments);
async function appoint(){
  if(busy.value)return;
  error.value='';success.value='';
  if(!clients.value.some(c=>c.id===clientId.value)||!specialization.value.trim()||experience.value===null||!Number.isInteger(experience.value)||experience.value<0){error.value='Выберите клиента, укажите специализацию и целое число лет опыта.';return;}
  busy.value=true;
  try{
    await adminApi.appointInstructor({user_id:clientId.value,specialization:specialization.value.trim(),experience_years:experience.value,...(description.value.trim()?{description:description.value.trim()}:{})});
    clientId.value=0;specialization.value='';experience.value=null;description.value='';appointing.value=false;
    await load();success.value='Инструктор назначен. Пользователю необходимо выйти и войти заново.';
  }catch(e){const msg=e instanceof Error?e.message:'Не удалось назначить инструктора';await load();error.value=msg;}
  finally{busy.value=false;}
}
async function changeAssignment(id:number,remove:boolean){
  if(busy.value||!current.value||!id)return;
  if(remove&&!confirm('Снять назначение услуги? Существующие бронирования сохранятся.'))return;
  busy.value=true;error.value='';success.value='';
  try{if(remove)await adminApi.removeService(selected.value,id);else await adminApi.assignService(selected.value,id);await loadAssignments();success.value=remove?'Назначение снято':'Услуга назначена';}
  catch(e){const msg=e instanceof Error?e.message:'Не удалось изменить назначение';await loadAssignments();error.value=msg;}
  finally{busy.value=false;}
}
onMounted(load);
</script>
<template>
  <div class="form-stack">
    <h2 class="h3">Инструкторы</h2>
    <button type="button" class="btn btn--primary" :disabled="busy||loading" @click="appointing=!appointing">{{appointing?'Закрыть форму':'Назначить инструктора из клиентов'}}</button>
    <form v-if="appointing" class="form-stack" @submit.prevent="appoint">
      <p v-if="!clients.length" class="muted">Нет клиентов, доступных для назначения.</p>
      <label>Клиент<select v-model.number="clientId" class="field" required :disabled="busy"><option :value="0" disabled>Выберите клиента</option><option v-for="c in clients" :key="c.id" :value="c.id">{{c.name}} — {{c.email}} · {{c.phone}}</option></select></label>
      <label>Специализация<input v-model.trim="specialization" class="field" required :disabled="busy"></label>
      <label>Опыт (лет)<input v-model.number="experience" class="field" type="number" min="0" step="1" required :disabled="busy"></label>
      <label>Описание (необязательно)<textarea v-model="description" class="field" :disabled="busy"></textarea></label>
      <button class="btn btn--primary" :disabled="busy||!clients.length">Назначить инструктором</button>
    </form>
    <p v-if="loading">Загрузка...</p>
    <p v-if="error" class="state-error" role="alert">{{error}}</p>
    <p v-if="success" role="status">{{success}}</p>
    <button type="button" class="btn btn--outline btn--sm" :disabled="busy||loading" @click="load">Обновить</button>
    <div v-if="!loading" class="table-wrap">
      <table class="table"><thead><tr><th>Имя</th><th>Специализация</th><th>Опыт</th><th>Статус</th></tr></thead><tbody>
        <tr v-for="i in instructors" :key="i.id"><td>{{i.name}}</td><td>{{i.specialization}}</td><td>{{i.experienceYears}} лет</td><td>{{i.isActive?'Активен':'Неактивен'}}</td></tr>
        <tr v-if="!instructors.length"><td colspan="4">Инструкторов пока нет.</td></tr>
      </tbody></table>
    </div>
    <template v-if="instructors.length">
      <h3 class="h3">Назначение услуг</h3>
      <label>Инструктор<select v-model.number="selected" class="field" :disabled="busy"><option v-for="i in instructors" :key="i.id" :value="i.id">{{i.name}}</option></select></label>
      <p v-if="!assigned.length" class="muted">Нет назначенных услуг.</p>
      <div v-for="s in assigned" :key="s.id" class="booking-line"><span>{{s.title}} {{s.active?'':'(неактивна)'}}</span><button type="button" class="btn btn--outline btn--sm" :disabled="busy" @click="changeAssignment(s.id,true)">Снять назначение</button></div>
      <label>Добавить услугу<select v-model.number="serviceId" class="field" :disabled="busy"><option :value="0">Выберите услугу</option><option v-for="s in availableServices" :key="s.id" :value="s.id">{{s.title}} {{s.active?'':'(неактивна)'}}</option></select></label>
      <button type="button" class="btn btn--primary" :disabled="busy||!serviceId" @click="changeAssignment(serviceId,false)">Назначить услугу</button>
    </template>
  </div>
</template>
