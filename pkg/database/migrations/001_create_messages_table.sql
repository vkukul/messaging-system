-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    "to" VARCHAR(255) NOT NULL,
    content VARCHAR(160) NOT NULL,
    sent BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP,
    message_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
