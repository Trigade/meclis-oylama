CREATE TABLE IF NOT EXISTS komisyonlar (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    aciklama    TEXT,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS komisyon_uyeler (
    komisyon_id INTEGER NOT NULL REFERENCES komisyonlar(id),
    member_id   INTEGER NOT NULL REFERENCES members(id),
    atandi_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (komisyon_id, member_id)
);

CREATE TABLE IF NOT EXISTS komisyon_kararlar (
    id          SERIAL PRIMARY KEY,
    komisyon_id INTEGER NOT NULL REFERENCES komisyonlar(id),
    karar_metni TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);