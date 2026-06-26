CREATE TABLE IF NOT EXISTS seats (
    id         SERIAL PRIMARY KEY,
    seat_no    INTEGER NOT NULL UNIQUE,
    member_id  INTEGER REFERENCES members(id),
    block_name VARCHAR(50)
);

INSERT INTO seats (seat_no, block_name)
SELECT i, CASE
    WHEN i BETWEEN 1  AND 8  THEN 'sol_ust'
    WHEN i BETWEEN 9  AND 18 THEN 'sol_alt'
    WHEN i BETWEEN 19 AND 26 THEN 'sag_1'
    WHEN i BETWEEN 27 AND 36 THEN 'sag_2'
    WHEN i BETWEEN 37 AND 46 THEN 'sag_3'
END
FROM generate_series(1, 46) AS i
ON CONFLICT DO NOTHING;

UPDATE seats s SET member_id = m.id
FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY id) as rn FROM members WHERE role = 'member') m
WHERE s.seat_no = m.rn;