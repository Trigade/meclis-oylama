CREATE TABLE IF NOT EXISTS attendance_sessions (
    id         SERIAL PRIMARY KEY,
    meeting_id VARCHAR(50)  NOT NULL,
    member_id  INTEGER      NOT NULL REFERENCES members(id),
    entered_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    exited_at  TIMESTAMPTZ,
    UNIQUE(meeting_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_attendance_meeting
    ON attendance_sessions(meeting_id);