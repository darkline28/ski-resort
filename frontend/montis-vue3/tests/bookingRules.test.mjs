import { test } from 'node:test';
import assert from 'node:assert/strict';
import { validRentalPeriod, validQuantity, canCancelBooking } from '../src/bookingRules.ts';

const now=Date.parse('2026-11-20T10:00:00+03:00');
test('Rental period: future, at least one full hour',()=>{
  assert.equal(validRentalPeriod('2026-11-20T11:00','2026-11-20T13:00',now),true);
  assert.equal(validRentalPeriod('2026-11-20T09:00','2026-11-20T11:00',now),false);
  assert.equal(validRentalPeriod('2026-11-20T11:00','2026-11-20T11:30',now),false);
  assert.equal(validRentalPeriod('2026-11-20T11:00','2026-11-20T12:30',now),false);
  assert.equal(validRentalPeriod('2026-11-20T11:00','2026-11-20T10:00',now),false);
  assert.equal(validRentalPeriod('','',now),false);
});
test('Quantity: positive integer within current availability',()=>{
  assert.equal(validQuantity(2,2),true);
  for(const n of [0,-1,1.5,3,NaN]) assert.equal(validQuantity(n,2),false);
  assert.equal(validQuantity(1,0),false);
});
test('Cancellation: only before start and from permitted status',()=>{
  const start='2026-11-20T11:00:00+03:00';
  assert.equal(canCancelBooking('lesson','confirmed',start,now),true);
  assert.equal(canCancelBooking('rental','pending',start,now),true);
  assert.equal(canCancelBooking('lesson','pending',start,now),false);
  for(const s of ['cancelled','issued','returned','completed','no_show']) assert.equal(canCancelBooking('rental',s,start,now),false);
  assert.equal(canCancelBooking('lesson','confirmed',start,Date.parse(start)),false);
});
