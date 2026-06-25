CREATE TABLE IF NOT EXISTS members (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    tc_no      VARCHAR(11)  NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role       VARCHAR(20)  NOT NULL DEFAULT 'member',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO members (name, tc_no, password_hash, role) VALUES
    ('Test Moderatör', '10000000000', '$2a$10$bdYsqhyqMEgagrkbks7uV./Ag6Yeu9NTs/ETmrPpoq2hxBUhDGS1q', 'moderator')
ON CONFLICT DO NOTHING;

INSERT INTO members (name, tc_no, password_hash, role)
SELECT 
  'Üye ' || i,
  '1000000' || LPAD(i::text, 4, '0'),
  '$2a$10$5OK320wCLl0vkcRoyRKZ/uIpRNonSqY/IidNui9VLox9IfnMQQKRu',
  'member'
FROM generate_series(1, 32) AS i
ON CONFLICT DO NOTHING;