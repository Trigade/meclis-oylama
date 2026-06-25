-- Roller ve katsayılar
CREATE TABLE IF NOT EXISTS member_roles (
    id         SERIAL PRIMARY KEY,
    role_name  VARCHAR(50) NOT NULL UNIQUE,
    katsayi    NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO member_roles (role_name, katsayi) VALUES
    ('Başkan', 2.0),
    ('Başkan Vekili', 1.5),
    ('Katip Üye', 1.2),
    ('Üye', 1.0)
ON CONFLICT DO NOTHING;

-- Üye rolü ataması
ALTER TABLE members
    ADD COLUMN IF NOT EXISTS member_role_id INTEGER REFERENCES member_roles(id);

-- Varsayılan olarak herkese "Üye" rolü ata
UPDATE members SET member_role_id = (
    SELECT id FROM member_roles WHERE role_name = 'Üye'
) WHERE member_role_id IS NULL;

-- Oturum huzur hakkı ayarları
CREATE TABLE IF NOT EXISTS huzur_hakki_settings (
    id          SERIAL PRIMARY KEY,
    meeting_id  VARCHAR(50) NOT NULL UNIQUE,
    taban_tutar NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
