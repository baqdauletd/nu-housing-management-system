CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

INSERT INTO roles (name) VALUES 
('student'), 
('housing'), 
('admin');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nu_id VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(120) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role_id INT REFERENCES roles(id),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE applications (
    id SERIAL PRIMARY KEY,
    student_id INT REFERENCES users(id) ON DELETE CASCADE,
    applicant_type VARCHAR(40),
    student_number VARCHAR(40),
    name_surname VARCHAR(255),
    fio VARCHAR(255),
    birth_date DATE,
    iin VARCHAR(20),
    school VARCHAR(255),
    level VARCHAR(80),
    passport_number VARCHAR(80),
    comments TEXT,
    year INT NOT NULL,
    major VARCHAR(120) NOT NULL,
    gender VARCHAR(10) NOT NULL,
    room_preference VARCHAR(50),
    additional_info TEXT,
    
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, canceled
    submitted_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    rejected_reason TEXT,
    reviewed_by INT REFERENCES users(id),
    review_timestamp TIMESTAMP
);

CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    application_id INT REFERENCES applications(id) ON DELETE CASCADE,
    type VARCHAR(80) NOT NULL,
    file_url TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE document_analysis (
    id SERIAL PRIMARY KEY,
    document_id INT UNIQUE REFERENCES documents(id) ON DELETE CASCADE,
    application_id INT REFERENCES applications(id) ON DELETE CASCADE,
    expected_type VARCHAR(80) NOT NULL,
    detected_category VARCHAR(40) NOT NULL,
    status VARCHAR(30) NOT NULL,
    has_astana_property BOOLEAN NOT NULL DEFAULT FALSE,
    has_astana_residence BOOLEAN NOT NULL DEFAULT FALSE,
    has_astana_employment BOOLEAN NOT NULL DEFAULT FALSE,
    issues_json JSONB NOT NULL DEFAULT '[]',
    reasoning_summary TEXT NOT NULL DEFAULT '',
    extracted_text_preview TEXT NOT NULL DEFAULT '',
    raw_ai_json TEXT NOT NULL DEFAULT '',
    analyzed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    read BOOLEAN DEFAULT FALSE
);


CREATE TABLE system_settings (
    id SERIAL PRIMARY KEY,
    applications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    application_open DATE,
    application_close DATE,
    required_documents JSONB DEFAULT '[]'
);

INSERT INTO system_settings (applications_enabled, application_open, application_close, required_documents)
VALUES (TRUE, '2025-01-01', '2025-12-31', '["id_card", "enrollment_certificate"]');

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    actor_id INT REFERENCES users(id),
    action VARCHAR(255),
    entity VARCHAR(50),
    entity_id INT,
    timestamp TIMESTAMP DEFAULT NOW()
);
