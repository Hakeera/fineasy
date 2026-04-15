


CREATE TABLE emails (
    id SERIAL PRIMARY KEY,
    gmail_id TEXT UNIQUE NOT NULL,
    subject TEXT,
    from_email TEXT,
    received_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE attachments (
    id SERIAL PRIMARY KEY,
    email_id INTEGER REFERENCES emails(id) ON DELETE CASCADE,
    filename TEXT,
    mime_type TEXT,
    file_path TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
