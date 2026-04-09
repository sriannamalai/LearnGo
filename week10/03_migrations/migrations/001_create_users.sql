-- ========================================
-- Migration 001: Create Users Table
-- ========================================
-- This is the initial migration that creates the users table.
-- Migrations are applied in order by version number.

CREATE TABLE users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(50) UNIQUE NOT NULL,
    full_name  VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create an index on username for fast lookups
CREATE INDEX idx_users_username ON users (username);

-- Insert a default admin user
INSERT INTO users (username, full_name) VALUES ('admin', 'System Administrator');
