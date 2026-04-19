BEGIN;

WITH student_role AS (
    SELECT id AS role_id FROM roles WHERE name = 'student'
),
seed AS (
    SELECT
        row_number() OVER (ORDER BY first_ord, last_ord)::int AS n,
        first_en,
        last_en,
        first_kk,
        last_kk
    FROM unnest(ARRAY[
        'Aruzhan', 'Dias', 'Aigerim', 'Alikhan', 'Madina',
        'Nursultan', 'Amina', 'Erasyl', 'Dana', 'Miras'
    ], ARRAY[
        'Аружан', 'Диас', 'Айгерім', 'Әлихан', 'Мәдина',
        'Нұрсұлтан', 'Әмина', 'Ерасыл', 'Дана', 'Мирас'
    ]) WITH ORDINALITY AS f(first_en, first_kk, first_ord)
    CROSS JOIN unnest(ARRAY[
        'Aitbayev', 'Sarsenova', 'Nurgaliyev', 'Tulegenova', 'Kassymov',
        'Ospanova', 'Bekzhanov', 'Serikova', 'Akhmetov', 'Iskakova'
    ], ARRAY[
        'Айтбаев', 'Сәрсенова', 'Нұрғалиев', 'Төлегенова', 'Қасымов',
        'Оспанова', 'Бекжанов', 'Серікова', 'Ахметов', 'Ысқақова'
    ]) WITH ORDINALITY AS l(last_en, last_kk, last_ord)
),
prepared AS (
    SELECT
        n,
        first_en,
        last_en,
        first_kk,
        last_kk,
        lower(first_en || '.' || last_en || '@nu.edu.kz') AS email,
        '+7701' || lpad((3000000 + n * 137)::text, 7, '0') AS phone,
        CASE WHEN n <= 70 THEN 'approved' ELSE 'rejected' END AS status,
        CASE WHEN n <= 70 THEN NULL ELSE 'Rejected test application' END AS rejected_reason,
        CASE WHEN n % 2 = 0 THEN 'female' ELSE 'male' END AS gender,
        CASE ((n - 1) % 5)
            WHEN 0 THEN 'Computer Science'
            WHEN 1 THEN 'Economics'
            WHEN 2 THEN 'Mechanical Engineering'
            WHEN 3 THEN 'Political Science'
            ELSE 'Biological Sciences'
        END AS major,
        1 + ((n - 1) % 4) AS study_year,
        '900' || lpad(((n * 3571 + 22222) % 1000000000)::text, 9, '0') AS iin,
        'N' || lpad(n::text, 7, '0') AS passport_number
    FROM seed
),
available_seed_nu_ids AS (
    SELECT
        row_number() OVER (ORDER BY candidate.nu_id) AS n,
        candidate.nu_id
    FROM (
        SELECT '99' || lpad(gs::text, 7, '0') AS nu_id
        FROM generate_series(1, 10000) AS gs
    ) AS candidate
    WHERE NOT EXISTS (
        SELECT 1
        FROM users u
        WHERE u.nu_id = candidate.nu_id
    )
    LIMIT 100
),
prepared_with_nu_ids AS (
    SELECT
        p.*,
        COALESCE(existing_user.nu_id, available.nu_id) AS nu_id
    FROM prepared p
    LEFT JOIN users existing_user
        ON existing_user.email = p.email
    LEFT JOIN available_seed_nu_ids available
        ON available.n = p.n
),
upserted_users AS (
    INSERT INTO users (nu_id, email, password_hash, role_id, phone, created_at, updated_at)
    SELECT
        p.nu_id,
        p.email,
        '',
        sr.role_id,
        p.phone,
        NOW(),
        NOW()
    FROM prepared_with_nu_ids p
    CROSS JOIN student_role sr
    ON CONFLICT (email) DO UPDATE
    SET
        phone = EXCLUDED.phone,
        role_id = EXCLUDED.role_id,
        updated_at = NOW()
    RETURNING id, email
),
prepared_with_users AS (
    SELECT
        p.*,
        uu.id AS user_id
    FROM prepared_with_nu_ids p
    JOIN upserted_users uu ON uu.email = p.email
),
paired AS (
    SELECT
        p.*,
        roommate.user_id AS roommate_user_id
    FROM prepared_with_users p
    LEFT JOIN prepared_with_users roommate
        ON p.n <= 50
       AND roommate.n = CASE WHEN p.n % 2 = 1 THEN p.n + 1 ELSE p.n - 1 END
),
deleted_existing_applications AS (
    DELETE FROM applications a
    USING prepared_with_users p
    WHERE a.student_id = p.user_id
    RETURNING a.id
)
INSERT INTO applications (
    student_id,
    applicant_type,
    student_number,
    name_surname,
    fio,
    birth_date,
    iin,
    school,
    level,
    passport_number,
    comments,
    year,
    major,
    gender,
    room_preference,
    additional_info,
    status,
    rejected_reason,
    submitted_at,
    updated_at,
    review_timestamp
)
SELECT
    p.user_id,
    'local',
    p.nu_id,
    p.first_kk || ' ' || p.last_kk,
    p.first_en || ' ' || p.last_en,
    (DATE '2000-01-01' + (p.n % 1800))::date,
    p.iin,
    'School of Sciences and Humanities',
    'Undergraduate',
    p.passport_number,
    'Seeded test student',
    p.study_year,
    p.major,
    p.gender,
    CASE WHEN p.n % 3 = 0 THEN 'single' ELSE 'double' END,
    concat_ws(E'\n',
        'first_name_en: ' || p.first_en,
        'last_name_en: ' || p.last_en,
        'first_name_kk: ' || p.first_kk,
        'last_name_kk: ' || p.last_kk,
        CASE WHEN p.roommate_user_id IS NOT NULL THEN 'preferred_roommate: ' || p.roommate_user_id::text END
    ),
    p.status,
    p.rejected_reason,
    NOW(),
    NOW(),
    CASE WHEN p.status IN ('approved', 'rejected') THEN NOW() ELSE NULL END
FROM paired p
ORDER BY p.n;

COMMIT;
