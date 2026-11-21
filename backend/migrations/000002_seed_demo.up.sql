-- Demonstration data for a fresh local installation.
-- Designed for the schema in 000001_init.up.sql.
-- Demo instructor accounts are intentionally unable to log in.

-- Services
INSERT INTO services (name, description, duration_minutes, price, is_active)
SELECT v.name, v.description, v.duration_minutes, v.price, TRUE
FROM (VALUES
 ('Демо: Индивидуальное занятие на лыжах', 'Занятие с инструктором для начинающих и продолжающих', 60, 2500),
 ('Демо: Индивидуальное занятие на сноуборде', 'Обучение базовой технике катания', 60, 2700),
 ('Демо: Детское занятие на лыжах', 'Занятие для детей с инструктором', 45, 2000)
) AS v(name, description, duration_minutes, price)
WHERE NOT EXISTS (SELECT 1 FROM services s WHERE s.name = v.name);

-- Categories
INSERT INTO equipment_categories (name, code)
VALUES
 ('Демо: Лыжи', 'demo_skis'),
 ('Демо: Сноуборды', 'demo_snowboards'),
 ('Демо: Лыжные ботинки', 'demo_ski_boots'),
 ('Демо: Ботинки для сноуборда', 'demo_snowboard_boots'),
 ('Демо: Шлемы', 'demo_helmets')
ON CONFLICT DO NOTHING;

-- Equipment catalogue (ski poles are included in skis by default)
INSERT INTO equipment (category_id, name, description, image_url, price_per_hour)
SELECT c.id, v.name, v.description, NULL, v.price
FROM (VALUES
 ('demo_skis', 'Демо: Горные лыжи Atomic', 'Горные лыжи с палками в комплекте', 650),
 ('demo_skis', 'Демо: Горные лыжи Rossignol', 'Горные лыжи с палками в комплекте', 600),
 ('demo_snowboards', 'Демо: Сноуборд Burton', 'Универсальный сноуборд', 700),
 ('demo_ski_boots', 'Демо: Лыжные ботинки Salomon', 'Ботинки для горных лыж', 350),
 ('demo_snowboard_boots', 'Демо: Ботинки для сноуборда', 'Ботинки для катания на сноуборде', 350),
 ('demo_helmets', 'Демо: Защитный шлем', 'Шлем для горных лыж и сноуборда', 200)
) AS v(code, name, description, price)
JOIN equipment_categories c ON c.code = v.code
WHERE NOT EXISTS (SELECT 1 FROM equipment e WHERE e.name = v.name);

-- Physical inventory: two units for each demo catalogue item
INSERT INTO equipment_units (equipment_id, inventory_number, size, status)
SELECT e.id, v.inventory_number, v.size, 'available'
FROM (VALUES
 ('Демо: Горные лыжи Atomic', 'DEMO-SKI-001', '165'),
 ('Демо: Горные лыжи Atomic', 'DEMO-SKI-002', '175'),
 ('Демо: Горные лыжи Rossignol', 'DEMO-SKI-003', '160'),
 ('Демо: Горные лыжи Rossignol', 'DEMO-SKI-004', '170'),
 ('Демо: Сноуборд Burton', 'DEMO-SNB-001', '150'),
 ('Демо: Сноуборд Burton', 'DEMO-SNB-002', '158'),
 ('Демо: Лыжные ботинки Salomon', 'DEMO-SBT-001', '42'),
 ('Демо: Лыжные ботинки Salomon', 'DEMO-SBT-002', '44'),
 ('Демо: Ботинки для сноуборда', 'DEMO-NBT-001', '42'),
 ('Демо: Ботинки для сноуборда', 'DEMO-NBT-002', '44'),
 ('Демо: Защитный шлем', 'DEMO-HLM-001', 'M'),
 ('Демо: Защитный шлем', 'DEMO-HLM-002', 'L')
) AS v(equipment_name, inventory_number, size)
JOIN equipment e ON e.name = v.equipment_name
ON CONFLICT (inventory_number) DO NOTHING;

-- Demo accounts for instructors. Password hash is deliberately invalid,
-- so these users cannot authenticate with a known shared password.
INSERT INTO users (name, email, phone, password_hash, role)
VALUES
 ('Анна Демонстрационная', 'demo.instructor.anna@example.invalid', '+70000000001', '!disabled-demo-account!', 'instructor'),
 ('Максим Демонстрационный', 'demo.instructor.max@example.invalid', '+70000000002', '!disabled-demo-account!', 'instructor')
ON CONFLICT (email) DO NOTHING;

INSERT INTO instructors (user_id, image_url, specialization, description, experience_years, is_active)
SELECT u.id, NULL, v.specialization, v.description, v.years, TRUE
FROM (VALUES
 ('demo.instructor.anna@example.invalid', 'Горные лыжи', 'Индивидуальные и детские занятия', 6),
 ('demo.instructor.max@example.invalid', 'Сноуборд', 'Обучение начинающих сноубордистов', 4)
) AS v(email, specialization, description, years)
JOIN users u ON u.email = v.email
ON CONFLICT (user_id) DO NOTHING;

-- Supported lesson types for each demo instructor
INSERT INTO instructor_services (instructor_id, service_id)
SELECT i.id, s.id
FROM instructors i
JOIN users u ON u.id = i.user_id
JOIN services s ON
 (u.email = 'demo.instructor.anna@example.invalid' AND s.name IN ('Демо: Индивидуальное занятие на лыжах', 'Демо: Детское занятие на лыжах'))
 OR (u.email = 'demo.instructor.max@example.invalid' AND s.name = 'Демо: Индивидуальное занятие на сноуборде')
ON CONFLICT DO NOTHING;

-- One future slot per instructor; never overlaps existing schedule for same instructor
INSERT INTO instructor_schedules (instructor_id, starts_at, ends_at)
SELECT i.id, v.start_at, v.start_at + INTERVAL '3 hours'
FROM instructors i
JOIN users u ON u.id = i.user_id
CROSS JOIN LATERAL (SELECT date_trunc('day', now()) + INTERVAL '2 days 10 hours' AS start_at) v
WHERE u.email IN ('demo.instructor.anna@example.invalid', 'demo.instructor.max@example.invalid')
AND NOT EXISTS (
  SELECT 1 FROM instructor_schedules x
  WHERE x.instructor_id = i.id
    AND tstzrange(x.starts_at, x.ends_at, '[)') && tstzrange(v.start_at, v.start_at + INTERVAL '3 hours', '[)')
);

-- Sets
INSERT INTO equipment_sets (name, description, sport_type, age_group, price_per_hour, is_active)
SELECT v.name, v.description, v.sport_type, 'adult', v.price, TRUE
FROM (VALUES
 ('Демо: Лыжный комплект', 'Лыжи (палки включены), ботинки и шлем', 'ski', 1000),
 ('Демо: Сноубордический комплект', 'Сноуборд, ботинки и шлем', 'snowboard', 1100)
) AS v(name, description, sport_type, price)
WHERE NOT EXISTS (SELECT 1 FROM equipment_sets es WHERE es.name = v.name);

INSERT INTO equipment_set_items (equipment_set_id, equipment_category_id, quantity)
SELECT es.id, c.id, 1
FROM (VALUES
 ('Демо: Лыжный комплект', 'demo_skis'),
 ('Демо: Лыжный комплект', 'demo_ski_boots'),
 ('Демо: Лыжный комплект', 'demo_helmets'),
 ('Демо: Сноубордический комплект', 'demo_snowboards'),
 ('Демо: Сноубордический комплект', 'demo_snowboard_boots'),
 ('Демо: Сноубордический комплект', 'demo_helmets')
) AS v(set_name, category_code)
JOIN equipment_sets es ON es.name = v.set_name
JOIN equipment_categories c ON c.code = v.category_code
ON CONFLICT DO NOTHING;
