import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'vite';

test('API integration contract: booking, cancellation, admin assignment and status', async () => {
  const vite=await createServer({server:{middlewareMode:true},appType:'custom',optimizeDeps:{noDiscovery:true,include:[]}});
  const originalFetch=globalThis.fetch;
  const calls=[];
  try {
    globalThis.fetch=async(url,opts)=>{
      calls.push({url,method:opts.method,body:opts.body?JSON.parse(opts.body):undefined});
      if (url.endsWith('/admin/clients')) return new Response(JSON.stringify([{id:7,name:'Анна',email:'a@example.test',phone:'+79990000000'}]));
      if (url.includes('/equipment/availability')) return new Response(JSON.stringify([{equipment_id:4,size:'M',available_quantity:1}]));
      return new Response('{}',{status:200});
    };
    const {catalogApi,bookingApi,adminApi,instructorApi}=await vite.ssrLoadModule('/src/api/index.ts');
    const start='2026-12-01T10:00:00+03:00',end='2026-12-01T12:00:00+03:00';
    assert.equal((await catalogApi.equipmentAvailability(4,start,end))[0].available_quantity,1);
    assert.equal(new URL(calls.at(-1).url,'http://localhost').searchParams.get('starts_at'),start);
    await bookingApi.createRental({equipmentId:4,size:'M',quantity:1,start,end});
    assert.deepEqual(calls.at(-1).body,{starts_at:start,ends_at:end,items:[{equipment_id:4,size:'M',quantity:1}],sets:[]});
    await bookingApi.createSet(8,1,[{category_id:2,size:'M'}],start,end);
    assert.deepEqual(calls.at(-1).body.sets,[{equipment_set_id:8,quantity:1,sizes:[{category_id:2,size:'M'}]}]);
    await bookingApi.cancel(4,'rental');
    assert.match(calls.at(-1).url,/\/equipment-bookings\/4\/cancel$/);
    assert.equal(calls.at(-1).method,'PATCH');
    assert.equal(calls.at(-1).body,undefined);
    assert.equal((await adminApi.clients())[0].id,7);
    await adminApi.appointInstructor({user_id:7,specialization:'Лыжи',experience_years:0});
    assert.match(calls.at(-1).url,/\/admin\/instructors$/);
    assert.deepEqual(calls.at(-1).body,{user_id:7,specialization:'Лыжи',experience_years:0});
    await adminApi.assignService(8,2);
    assert.deepEqual(calls.at(-1).body,{service_id:2});
    await instructorApi.setStatus(9,'no_show');
    assert.match(calls.at(-1).url,/\/instructors\/me\/bookings\/9\/status$/);
    assert.deepEqual(calls.at(-1).body,{status:'no_show'});
  } finally {
    globalThis.fetch=originalFetch;
    await vite.close();
  }
});
