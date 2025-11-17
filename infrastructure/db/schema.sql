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
    uploaded_at TIMESTAMP DEFAULT NOW()
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
    application_open DATE,
    application_close DATE,
    required_documents JSONB DEFAULT '[]'
);

INSERT INTO system_settings (application_open, application_close, required_documents)
VALUES ('2025-01-01', '2025-12-31', '["id_card", "enrollment_certificate"]');

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    actor_id INT REFERENCES users(id),
    action VARCHAR(255),
    entity VARCHAR(50),
    entity_id INT,
    timestamp TIMESTAMP DEFAULT NOW()
);
