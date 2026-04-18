BEGIN;

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS content_type VARCHAR(255);

UPDATE documents
SET original_filename = COALESCE(original_filename, split_part(file_url, '/', array_length(string_to_array(file_url, '/'), 1))),
    content_type = COALESCE(content_type, 'application/pdf')
WHERE original_filename IS NULL
   OR content_type IS NULL;

ALTER TABLE documents
    ALTER COLUMN original_filename SET NOT NULL,
    ALTER COLUMN content_type SET NOT NULL;

CREATE TABLE IF NOT EXISTS document_analysis (
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

CREATE INDEX IF NOT EXISTS idx_document_analysis_application_id
    ON document_analysis (application_id);

COMMIT;
