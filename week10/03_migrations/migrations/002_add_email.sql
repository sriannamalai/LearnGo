-- ========================================
-- Migration 002: Add Email Column to Users
-- ========================================
-- This migration adds an email column to the users table.
-- It demonstrates how to evolve a schema over time.

-- Add the email column (nullable at first, so existing rows aren't broken)
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- Create a unique index on email
CREATE UNIQUE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;

-- Update the admin user with an email
UPDATE users SET email = 'admin@example.com' WHERE username = 'admin';

-- Add an active flag with a default value
ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT true;
