<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { adminApi, catalogApi } from '@/api';

type Field = { key: string; label: string; type?: string; required?: boolean; choices?: string[]; min?: number; createOnly?: boolean };
const resources: Record<string, {title:string; fields:Field[]}> = {
  services:{title:'Услуги',fields:[{key:'name',label:'Название',required:true},{key:'duration_minutes',label:'Длительность (мин.)',type:'number',min:1,required:true},{key:'price',label:'Цена',type:'number',min:0,required:true},{key:'description',label:'Описание'},{key:'is_active',label:'Активность',choices:['true','false']}]},
  instructors:{title:'Редактирование инструкторов',fields:[{key:'specialization',label:'Специализация',required:true},{key:'experience_years',label:'Стаж (лет)',type:'number',min:0,required:true},{key:'description',label:'Описание'},{key:'image_url',label:'Ссылка на фото'},{key:'is_active',label:'Активность',choices:['true','false']}]},
  equipment:{title:'Оборудование',fields:[{key:'category_id',label:'ID категории',type:'number',min:1,required:true},{key:'name',label:'Название',required:true},{key:'price_per_hour',label:'Цена за час',type:'number',min:0,required:true},{key:'description',label:'Описание'},{key:'image_url',label:'Ссылка на фото'}]},
  'equipment/units':{title:'Экземпляры оборудования',fields:[{key:'equipment_id',label:'ID модели',type:'number',min:1,required:true,createOnly:true},{key:'inventory_number',label:'Инвентарный номер',required:true},{key:'size',label:'Размер',required:true},{key:'status',label:'Статус',choices:['available','maintenance','written_off'],required:true}]},
  'equipment/sets':{title:'Комплекты',fields:[{key:'name',label:'Название',required:true},{key:'sport_type',label:'Вид спорта',choices:['ski','snowboard'],required:true},{key:'age_group',label:'Возраст',choices:['adult','child'],required:true},{key:'price_per_hour',label:'Цена за час',type:'number',min:0,required:true},{key:'items',label:'Состав: ID категории и количество через двоеточие, позиции через запятую (1:1,5:1)',required:true},{key:'description',label:'Описание'},{key:'is_active',label:'Активность',choices:['true','false']}]},
  schedules:{title:'Расписание',fields:[{key:'starts_at',label:'Начало (МСК)',type:'datetime-local',required:true},{key:'ends_at',label:'Конец (МСК)',type:'datetime-local',required:true}]}
};
const resource=ref('services'), id=ref<number|null>(null), instructorId=ref<number|null>(null), values=ref<Record<string,string>>({}), rows=ref<{id:number;name:string}[]>([]), error=ref(''), success=ref(''), busy=ref(false);
const fields=computed(()=>resources[resource.value].fields.filter(f=>!id.value || !f.createOnly));
const path=computed(()=>resource.value==='schedules'?`instructors/${instructorId.value}/schedules`:resource.value);
let loadId=0;
async function load(){
  const token=++loadId;
  rows.value=[]; error.value='';
  try {
    let result:{id:number;name:string}[]=[];
    if(resource.value==='services') result=(await adminApi.services()).map(x=>({id:x.id,name:x.title+(x.active?'':' (неактивна)')}));
    if(resource.value==='instructors') result=(await adminApi.instructors()).map(x=>({id:x.id,name:x.name}));
    if(resource.value==='equipment') result=(await catalogApi.equipment()).map(x=>({id:x.id,name:x.title}));
    if(resource.value==='equipment/sets') result=await catalogApi.equipmentSets();
    if(resource.value==='equipment/units') result=(await adminApi.units()).map(x=>({id:Number(x.id),name:String(x.inventory_number)}));
    if(resource.value==='schedules' && Number.isInteger(instructorId.value) && Number(instructorId.value)>0) result=(await adminApi.schedules(Number(instructorId.value))).map(x=>({id:Number(x.id),name:String(x.starts_at)+' — '+String(x.ends_at)}));
    if(token===loadId) rows.value=result;
  } catch(e){if(token===loadId) error.value=e instanceof Error?e.message:'Не удалось загрузить';}
}
watch([resource,instructorId],()=>{id.value=null;values.value={};success.value='';void load();},{immediate:true});
watch(id,()=>{values.value={};success.value='';});
async function submit(){
  if(busy.value)return;
  error.value='';success.value='';
  try {
    if(resource.value==='schedules' && (!Number.isInteger(instructorId.value)||Number(instructorId.value)<=0)) throw new Error('Укажите ID инструктора');
    if(resource.value==='instructors' && !id.value) throw new Error('Для назначения нового инструктора используйте вкладку «Инструкторы».');
    const body:Record<string,unknown>={};
    for(const f of fields.value){
      const v=String(values.value[f.key]??'').trim();
      if(!v){if(!id.value && f.required)throw new Error('Заполните поле: '+f.label);continue;}
      if(f.type==='number'){const n=Number(v);if(!Number.isInteger(n)||n<(f.min??0))throw new Error('Некорректное значение: '+f.label);body[f.key]=n;}
      else if(f.type==='datetime-local'){if(!Number.isFinite(Date.parse(v+':00+03:00')))throw new Error('Некорректная дата');body[f.key]=v+':00+03:00';}
      else if(f.key==='items'){const items=v.split(',').map(x=>{const [a,b,...rest]=x.split(':').map(Number);if(rest.length||!Number.isInteger(a)||a<1||!Number.isInteger(b)||b<1)throw new Error('Некорректный состав комплекта');return {category_id:a,quantity:b};});if(new Set(items.map(x=>x.category_id)).size!==items.length)throw new Error('Категории не должны повторяться');body.items=items;}
      else body[f.key]=f.key==='is_active'?v==='true':v;
    }
    if(!Object.keys(body).length)throw new Error('Укажите хотя бы одно поле');
    if(body.starts_at && body.ends_at && Date.parse(String(body.starts_at))>=Date.parse(String(body.ends_at)))throw new Error('Конец должен быть позже начала');
    busy.value=true;
    await adminApi.save(path.value,body,id.value??undefined);
    values.value={}; await load();success.value='Изменения сохранены';
  }catch(e){error.value=e instanceof Error?e.message:'Не удалось сохранить';}finally{busy.value=false;}
}
async function remove(){
  if(busy.value||!id.value||!confirm('Удалить или деактивировать выбранную запись?'))return;
  busy.value=true;error.value='';success.value='';
  try{await adminApi.remove(path.value,id.value);id.value=null;await load();success.value='Запись удалена или деактивирована';}
  catch(e){error.value=e instanceof Error?e.message:'Не удалось удалить';}finally{busy.value=false;}
}
</script>
<template>
  <form class="form-stack" @submit.prevent="submit">
    <label>Раздел<select v-model="resource" class="field" :disabled="busy"><option v-for="(r,key) in resources" :key="key" :value="key">{{r.title}}</option></select></label>
    <label v-if="resource==='schedules'">ID инструктора<input v-model.number="instructorId" class="field" type="number" min="1" step="1" :disabled="busy" required /></label>
    <label>Действие<select v-model="id" class="field" :disabled="busy"><option :value="null">{{resource==='instructors'?'Выберите инструктора':'Создать'}}</option><option v-for="r in rows" :key="r.id" :value="r.id">#{{r.id}} — {{r.name}}</option></select></label>
    <p v-if="id" class="muted">Заполните только поля, которые хотите изменить.</p>
    <label v-for="f in fields" :key="f.key">{{f.label}}
      <select v-if="f.choices" v-model="values[f.key]" class="field" :disabled="busy" :required="!id && f.required"><option value="">Не выбрано</option><option v-for="c in f.choices" :key="c" :value="c">{{c}}</option></select>
      <input v-else v-model="values[f.key]" class="field" :type="f.type||'text'" :min="f.min" :step="f.type==='number' ? 1 : undefined" :disabled="busy" :required="!id && f.required"/>
    </label>
    <p v-if="error" class="state-error">{{error}}</p><p v-if="success">{{success}}</p>
    <button class="btn btn--primary" :disabled="busy">Сохранить</button>
    <button v-if="id && resource!=='equipment/units'" type="button" class="btn btn--outline" :disabled="busy" @click="remove">Удалить / деактивировать</button>
    <button type="button" class="btn btn--outline" :disabled="busy" @click="load">Обновить список</button>
  </form>
</template>
