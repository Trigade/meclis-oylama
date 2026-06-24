CREATE TABLE IF NOT EXISTS members (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    tc_no      VARCHAR(11)  NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO members (name, tc_no, password_hash, role) VALUES
    ('Test Moderatör', '10000000000', '$2a$10$placeholder', 'moderator'),
    ('Üye 1',          '10000000001', '$2a$10$placeholder', 'member'),
    ('Üye 2',          '10000000002', '$2a$10$placeholder', 'member'),
    ('Üye 3',          '10000000003', '$2a$10$placeholder', 'member')
ON CONFLICT DO NOTHING;