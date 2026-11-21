
BEGIN;

-- 1. Убираем приставку "Демо:" у услуг
UPDATE services
SET name = regexp_replace(name, '^Демо: ', '')
WHERE name IN (
    'Демо: Индивидуальное занятие на лыжах',
    'Демо: Индивидуальное занятие на сноуборде',
    'Демо: Детское занятие на лыжах'
);

-- 2. Переносим оборудование и комплекты
-- в уже существующие обычные категории

DO $$
DECLARE
    r RECORD;
    target_id INTEGER;
BEGIN
    FOR r IN
        SELECT id, code FROM equipment_categories
        WHERE code IN (
            'demo_skis',
            'demo_snowboards',
            'demo_ski_boots',
            'demo_snowboard_boots',
            'demo_helmets'
        )
    LOOP
        SELECT id INTO target_id
        FROM equipment_categories
        WHERE name = CASE
            WHEN r.code = 'demo_skis' THEN 'Лыжи'
            WHEN r.code = 'demo_snowboards' THEN 'Сноуборды'
            WHEN r.code IN ('demo_ski_boots', 'demo_snowboard_boots')
                THEN 'Ботинки'
            WHEN r.code = 'demo_helmets' THEN 'Шлемы'
        END
        AND id <> r.id
        LIMIT 1;

        IF target_id IS NULL THEN
            RAISE EXCEPTION
                'Не найдена основная категория для %', r.code;
        END IF;

        UPDATE equipment
        SET category_id = target_id
        WHERE category_id = r.id;

        INSERT INTO equipment_set_items (
            equipment_set_id,
            equipment_category_id,
            quantity
        )
        SELECT equipment_set_id, target_id, quantity
        FROM equipment_set_items
        WHERE equipment_category_id = r.id
        ON CONFLICT (equipment_set_id, equipment_category_id)
        DO UPDATE SET quantity = EXCLUDED.quantity;

        DELETE FROM equipment_set_items
        WHERE equipment_category_id = r.id;

        DELETE FROM equipment_categories WHERE id = r.id;
    END LOOP;
END $$;

-- 3. Убираем приставку у оборудования
UPDATE equipment
SET name = regexp_replace(name, '^Демо: ', '')
WHERE name LIKE 'Демо: %';

-- 4. Переименовываем комплекты
UPDATE equipment_sets
SET name = regexp_replace(name, '^Демо: ', '')
WHERE name LIKE 'Демо: %';

-- 5. Нормальные имена инструкторов
UPDATE users
SET name = 'Анна Смирнова'
WHERE email = 'demo.instructor.anna@example.invalid';

UPDATE users
SET name = 'Максим Волков'
WHERE email = 'demo.instructor.max@example.invalid';

-- Назначаем рабочие пароли инструкторам
UPDATE users
SET password_hash = '$2a$10$nYIQ/KBIKsTCqGrZ.7b7ZuqwCMFTquqZFPUYMJnXou5XQc.5WioPW'
WHERE email IN (
    'demo.instructor.anna@example.invalid',
    'demo.instructor.max@example.invalid'
);

COMMIT;
