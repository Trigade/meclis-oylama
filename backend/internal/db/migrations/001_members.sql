CREATE TABLE IF NOT EXISTS members (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    tc_no      VARCHAR(11)  NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Moderatörü manuel ekleyelim
INSERT INTO members (name, tc_no, password_hash, role) 
VALUES ('Test Moderatör', '10000000000', '$2a$10$placeholder', 'moderator')
ON CONFLICT DO NOTHING;

-- 35 Tane normal üyeyi otomatik oluşturalım
INSERT INTO members (name, tc_no, password_hash, role)
SELECT 
    'Üye ' || i, 
    (10000000000 + i)::text, 
    '$2a$10$placeholder', 
    'member'
FROM generate_series(1, 35) AS i
ON CONFLICT DO NOTHING;