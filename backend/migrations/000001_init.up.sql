CREATE EXTENSION IF NOT EXISTS btree_gist;

create table users (
	id INTEGER generated always as identity primary key,
	name VARCHAR(100) not null,
	email VARCHAR(255) not null unique,
	phone VARCHAR(20),
	password_hash VARCHAR(255) not null,
	role VARCHAR(20) not null check (role in ('client', 'instructor', 'admin')),
	created_at TIMESTAMPTZ not null default NOW()
);

create table instructors (
	id INTEGER generated always as identity primary key,
	user_id INTEGER not null unique references users(id),
	specialization VARCHAR(200) not null,
	description TEXT,
	experience_years INTEGER NOT NULL CHECK (experience_years >= 0),
	is_active BOOLEAN not null default true
);

create table services(
	id INTEGER generated always as identity primary key,
	name VARCHAR(100) not null,
	description TEXT,
	duration_minutes INTEGER not null CHECK(duration_minutes > 0),
	price NUMERIC(10,2) not null CHECK(price >= 0),
	is_active BOOLEAN not null default true
);

create table instructor_services (
    instructor_id INTEGER not null references instructors(id),
    service_id INTEGER not null references services(id),

    primary key (instructor_id, service_id)
);

CREATE TABLE instructor_schedules (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    instructor_id INTEGER NOT NULL REFERENCES instructors(id),

    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,

    CHECK (ends_at > starts_at),
    
    CONSTRAINT instructor_schedules_no_overlap
        EXCLUDE USING gist (
            instructor_id WITH =,
            tstzrange(starts_at, ends_at, '[)') WITH &&
        )
);

CREATE TABLE lesson_bookings (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    instructor_id INTEGER NOT NULL REFERENCES instructors(id),
    service_id INTEGER NOT NULL REFERENCES services(id),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL
        CHECK (status IN (
            'confirmed',
            'cancelled',
            'completed',
            'no_show'
        )),
    total_price NUMERIC(10,2) NOT NULL
        CHECK (total_price >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at),

    CONSTRAINT lesson_no_overlap
        EXCLUDE USING gist (
            instructor_id WITH =,
            tstzrange(starts_at, ends_at, '[)') WITH &&
        )
    WHERE (status IN ('confirmed', 'completed', 'no_show'))
);

CREATE TABLE equipment_categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    price_per_hour NUMERIC(10,2) NOT NULL
        CHECK (price_per_hour >= 0)
);

CREATE TABLE equipment (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES equipment_categories(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    image_url VARCHAR(500)
);

CREATE TABLE equipment_units (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    equipment_id INTEGER NOT NULL REFERENCES equipment(id),
    inventory_number VARCHAR(80) NOT NULL UNIQUE,
    size VARCHAR(30) NOT NULL,
    status VARCHAR(20) NOT NULL
        CHECK (status IN (
            'available',
            'maintenance',
            'written_off'
        )),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE equipment_sets (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    sport_type VARCHAR(20) NOT NULL
        CHECK (sport_type IN ('ski', 'snowboard')),
    age_group VARCHAR(20) NOT NULL
        CHECK (age_group IN ('adult', 'child')),
    price_per_hour NUMERIC(10,2) NOT NULL
        CHECK (price_per_hour >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE equipment_set_items (
    equipment_set_id INTEGER NOT NULL
        REFERENCES equipment_sets(id),
    equipment_category_id INTEGER NOT NULL
        REFERENCES equipment_categories(id),
    quantity INTEGER NOT NULL
        CHECK (quantity > 0),
    PRIMARY KEY (
        equipment_set_id,
        equipment_category_id
    )
);

CREATE TABLE equipment_bookings (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL
        CHECK (status IN (
            'confirmed',
            'issued',
            'returned',
            'cancelled'
        )),
    total_price NUMERIC(10,2) NOT NULL
        CHECK (total_price >= 0),
    CHECK (ends_at > starts_at)
);

CREATE TABLE booked_equipment_sets (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    booking_id INTEGER NOT NULL
        REFERENCES equipment_bookings(id),
    equipment_set_id INTEGER NOT NULL
        REFERENCES equipment_sets(id),
    total_price NUMERIC(10,2) NOT NULL
        CHECK (total_price >= 0),
    UNIQUE (id, booking_id)
);

CREATE TABLE equipment_booking_items (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    booking_id INTEGER NOT NULL REFERENCES equipment_bookings(id),
    equipment_unit_id INTEGER NOT NULL REFERENCES equipment_units(id),
    booked_set_id INTEGER,
    item_total_price NUMERIC(10,2) CHECK (item_total_price >= 0),
    UNIQUE (booking_id, equipment_unit_id),
    FOREIGN KEY (booked_set_id, booking_id)
        REFERENCES booked_equipment_sets(id, booking_id),
    CHECK (
        (
            booked_set_id IS NULL
            AND item_total_price IS NOT NULL
        )
        OR
        (
            booked_set_id IS NOT NULL
            AND item_total_price IS NULL
        )
    )
);

CREATE INDEX idx_lesson_bookings_user_id
ON lesson_bookings(user_id);

CREATE INDEX idx_equipment_bookings_user_id
ON equipment_bookings(user_id);

CREATE INDEX idx_equipment_units_equipment_id
ON equipment_units(equipment_id);

CREATE INDEX idx_equipment_booking_items_unit_id
ON equipment_booking_items(equipment_unit_id);

CREATE INDEX idx_equipment_category_id
ON equipment(category_id)