-- migrations/000001_create_passkeeper_tables.up.sql

-- Create table for users
CREATE TABLE IF NOT EXISTS users ( 
 id UUID PRIMARY KEY,
 username TEXT NOT NULL UNIQUE,
 password_hash TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create table for items
CREATE TABLE IF NOT EXISTS items ( 
 id UUID PRIMARY KEY,
 user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 type SMALLINT NOT NULL CHECK (type in (0,1, 2, 3, 4)),
 data bytea NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 deleted_at TIMESTAMPTZ NULL
);

-- Create index for items
CREATE INDEX IF NOT EXISTS idx_items_user_updated_at ON items(user_id, updated_at);
CREATE INDEX IF NOT EXISTS idx_items_user_deleted_at ON items(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_items_user_type_updated_at ON items(user_id, type, updated_at);
    