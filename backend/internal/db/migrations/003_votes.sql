CREATE TABLE IF NOT EXISTS votings (
    id         SERIAL PRIMARY KEY,
    meeting_id VARCHAR(50)  NOT NULL,
    title      VARCHAR(255) NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'pending',
    result     VARCHAR(20),
    started_at TIMESTAMPTZ,
    ended_at   TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS votes (
    id        SERIAL PRIMARY KEY,
    voting_id INTEGER     NOT NULL REFERENCES votings(id),
    member_id INTEGER     NOT NULL REFERENCES members(id),
    choice    VARCHAR(10) NOT NULL CHECK (choice IN ('evet', 'hayir')),
    cast_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(voting_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_votes_voting ON votes(voting_id);