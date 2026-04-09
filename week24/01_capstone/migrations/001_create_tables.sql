-- ========================================
-- TaskFlow Database Schema
-- Migration: 001_create_tables.sql
-- ========================================
-- This migration creates the initial database schema for TaskFlow.
-- It demonstrates Week 10's PostgreSQL lessons:
--   - Table design with proper data types
--   - Indexes for common query patterns
--   - Foreign key constraints for referential integrity
--   - Timestamps with timezone support
--   - Array columns (tags)
--   - CHECK constraints for data validation
--
-- Run this migration:
--   taskflow migrate up
-- ========================================

-- ========================================
-- Enable UUID Extension
-- ========================================
-- We use UUIDs as primary keys for several reasons:
--   1. Globally unique — safe for distributed systems
--   2. No sequence contention under high concurrency
--   3. Can be generated client-side (offline support)
--   4. Don't leak information about record count
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ========================================
-- Users Table
-- ========================================
CREATE TABLE IF NOT EXISTS users (
    id            VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    email         VARCHAR(255) NOT NULL UNIQUE,
    name          VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'user'
                  CHECK (role IN ('user', 'admin')),
    active        BOOLEAN      NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);

-- Index for email lookups during authentication
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Index for listing active users
CREATE INDEX IF NOT EXISTS idx_users_active ON users (active) WHERE active = true;

-- ========================================
-- Tasks Table
-- ========================================
CREATE TABLE IF NOT EXISTS tasks (
    id            VARCHAR(64) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    user_id       VARCHAR(64)  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         VARCHAR(500) NOT NULL,
    description   TEXT         DEFAULT '',
    priority      VARCHAR(20)  NOT NULL DEFAULT 'medium'
                  CHECK (priority IN ('low', 'medium', 'high')),
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending'
                  CHECK (status IN ('pending', 'in_progress', 'completed', 'archived')),
    tags          TEXT[]       DEFAULT '{}',
    due_date      TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ
);

-- ========================================
-- Indexes for Common Query Patterns
-- ========================================

-- User's tasks (most common query — listing tasks for a user)
CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks (user_id);

-- Filtering by status (e.g., "show me all pending tasks")
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);

-- Filtering by priority
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks (priority);

-- Composite index for the most common query pattern:
-- "List pending tasks for user X, ordered by creation date"
CREATE INDEX IF NOT EXISTS idx_tasks_user_status_created
    ON tasks (user_id, status, created_at DESC);

-- GIN index for array contains queries on tags
-- This enables: WHERE 'urgent' = ANY(tags)
CREATE INDEX IF NOT EXISTS idx_tasks_tags ON tasks USING GIN (tags);

-- Partial index for overdue tasks (only indexes non-completed tasks with due dates)
-- Partial indexes are smaller and faster because they only include matching rows.
CREATE INDEX IF NOT EXISTS idx_tasks_overdue
    ON tasks (due_date)
    WHERE status NOT IN ('completed', 'archived') AND due_date IS NOT NULL;

-- ========================================
-- Updated-at Trigger
-- ========================================
-- Automatically update the updated_at timestamp when a row is modified.
-- This ensures the timestamp is always accurate, even if the application
-- forgets to set it.

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- Schema Migration Tracking
-- ========================================
-- This table is also created by the migration tool itself,
-- but we include it here for documentation completeness.
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
