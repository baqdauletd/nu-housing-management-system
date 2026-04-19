BEGIN;

CREATE TABLE IF NOT EXISTS room_allocations (
    id SERIAL PRIMARY KEY,
    application_id INT UNIQUE REFERENCES applications(id) ON DELETE CASCADE,
    student_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    block INT NOT NULL,
    room_number INT NOT NULL,
    bed_number INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(block, room_number, bed_number)
);

ALTER TABLE room_allocations ADD COLUMN IF NOT EXISTS ends_at TIMESTAMP NULL;
CREATE INDEX IF NOT EXISTS idx_room_allocations_student_id ON room_allocations(student_id);
CREATE INDEX IF NOT EXISTS idx_room_allocations_room ON room_allocations(block, room_number);

WITH seeded_approved AS (
    SELECT
        a.id AS application_id,
        a.student_id,
        a.gender,
        a.major,
        substring(a.additional_info from 'preferred_roommate: ([0-9]+)')::int AS preferred_student_id
    FROM applications a
    WHERE a.comments = 'Seeded test student'
      AND a.status = 'approved'
),
deleted_seeded_allocations AS (
    DELETE FROM room_allocations ra
    USING seeded_approved s
    WHERE ra.application_id = s.application_id
    RETURNING ra.id
),
mutual_preferred_pairs AS (
    SELECT
        LEAST(a.student_id, b.student_id) AS pair_key,
        a.application_id AS application_id,
        a.student_id AS student_id,
        row_number() OVER (
            PARTITION BY LEAST(a.student_id, b.student_id), GREATEST(a.student_id, b.student_id)
            ORDER BY a.student_id
        ) AS bed_number
    FROM seeded_approved a
    JOIN seeded_approved b
      ON b.student_id = a.preferred_student_id
     AND b.preferred_student_id = a.student_id
    WHERE a.preferred_student_id IS NOT NULL
),
preferred_groups AS (
    SELECT
        dense_rank() OVER (ORDER BY pair_key) AS group_number,
        application_id,
        student_id,
        bed_number
    FROM mutual_preferred_pairs
),
students_in_preferred_groups AS (
    SELECT student_id FROM preferred_groups
),
unpaired_approved AS (
    SELECT
        application_id,
        student_id,
        row_number() OVER (ORDER BY application_id) AS seq
    FROM seeded_approved
    WHERE student_id NOT IN (SELECT student_id FROM students_in_preferred_groups)
),
unpaired_groups AS (
    SELECT
        25 + ((seq + 1) / 2) AS group_number,
        application_id,
        student_id,
        CASE WHEN seq % 2 = 1 THEN 1 ELSE 2 END AS bed_number
    FROM unpaired_approved
),
allocation_groups AS (
    SELECT group_number, application_id, student_id, bed_number FROM preferred_groups
    UNION ALL
    SELECT group_number, application_id, student_id, bed_number FROM unpaired_groups
),
candidate_rooms AS (
    SELECT
        block,
        floor * 100 + room AS room_number,
        row_number() OVER (ORDER BY block, floor, room) AS room_seq
    FROM generate_series(22, 27) AS block
    CROSS JOIN generate_series(2, 12) AS floor
    CROSS JOIN generate_series(1, 28) AS room
),
empty_candidate_rooms AS (
    SELECT
        cr.block,
        cr.room_number,
        row_number() OVER (ORDER BY cr.block, cr.room_number) AS available_seq
    FROM candidate_rooms cr
    WHERE NOT EXISTS (
        SELECT 1
        FROM room_allocations ra
        WHERE ra.block = cr.block
          AND ra.room_number = cr.room_number
    )
),
assignments AS (
    SELECT
        ag.application_id,
        ag.student_id,
        er.block,
        er.room_number,
        ag.bed_number
    FROM allocation_groups ag
    JOIN empty_candidate_rooms er
      ON er.available_seq = ag.group_number
)
INSERT INTO room_allocations (application_id, student_id, block, room_number, bed_number, created_at)
SELECT
    application_id,
    student_id,
    block,
    room_number,
    bed_number,
    NOW()
FROM assignments
ORDER BY block, room_number, bed_number
ON CONFLICT (application_id) DO UPDATE
SET
    student_id = EXCLUDED.student_id,
    block = EXCLUDED.block,
    room_number = EXCLUDED.room_number,
    bed_number = EXCLUDED.bed_number,
    created_at = EXCLUDED.created_at;

COMMIT;
