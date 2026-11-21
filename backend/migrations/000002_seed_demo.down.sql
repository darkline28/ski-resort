-- Rollback of demo data. If demo instructors or equipment have been booked,
-- foreign-key constraints may prevent rollback; that is intentional to protect bookings.
DELETE FROM instructor_schedules WHERE instructor_id IN (
  SELECT i.id FROM instructors i JOIN users u ON u.id=i.user_id
  WHERE u.email IN ('demo.instructor.anna@example.invalid','demo.instructor.max@example.invalid')
);
DELETE FROM instructor_services WHERE instructor_id IN (
  SELECT i.id FROM instructors i JOIN users u ON u.id=i.user_id
  WHERE u.email IN ('demo.instructor.anna@example.invalid','demo.instructor.max@example.invalid')
);
DELETE FROM instructors WHERE user_id IN (
  SELECT id FROM users WHERE email IN ('demo.instructor.anna@example.invalid','demo.instructor.max@example.invalid')
);
DELETE FROM users WHERE email IN ('demo.instructor.anna@example.invalid','demo.instructor.max@example.invalid');
DELETE FROM equipment_set_items WHERE equipment_set_id IN (
  SELECT id FROM equipment_sets WHERE name IN ('Демо: Лыжный комплект','Демо: Сноубордический комплект')
);
DELETE FROM equipment_sets WHERE name IN ('Демо: Лыжный комплект','Демо: Сноубордический комплект');
DELETE FROM equipment_units WHERE inventory_number LIKE 'DEMO-%';
DELETE FROM equipment WHERE name IN (
 'Демо: Горные лыжи Atomic','Демо: Горные лыжи Rossignol','Демо: Сноуборд Burton',
 'Демо: Лыжные ботинки Salomon','Демо: Ботинки для сноуборда','Демо: Защитный шлем'
);
DELETE FROM equipment_categories WHERE code IN ('demo_skis','demo_snowboards','demo_ski_boots','demo_snowboard_boots','demo_helmets');
DELETE FROM services WHERE name IN (
 'Демо: Индивидуальное занятие на лыжах',
 'Демо: Индивидуальное занятие на сноуборде',
 'Демо: Детское занятие на лыжах'
);
