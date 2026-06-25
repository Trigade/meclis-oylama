CREATE TABLE IF NOT EXISTS meetings (
    id          SERIAL PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    meeting_no  VARCHAR(50),
    status      VARCHAR(20) NOT NULL DEFAULT 'planned',
    planned_at  TIMESTAMPTZ,
    started_at  TIMESTAMPTZ,
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_meetings_status ON meetings(status);