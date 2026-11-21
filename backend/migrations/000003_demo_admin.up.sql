
INSERT INTO users (
    name,
    email,
    phone,
    password_hash,
    role
)
VALUES (
    'Demo Admin',
    'admin@gmail.com',
    '+79990000000',
    '$2a$10$YFZiM.eOYd1BwOXBTwm7N.oRCkIck0xVnG2azPNRgZmX9YmhqP9ey',
    'admin'
)
ON CONFLICT (email) DO NOTHING;
