-- migrations/000001_create_passkeeper_tables.down.sql

-- Drop index for items
DROP INDEX IF EXISTS idx_items_user_updated_at;
DROP INDEX IF EXISTS idx_items_user_deleted_at;
DROP INDEX IF EXISTS idx_items_user_type_updated_at;

-- Drop tables
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS users;


-- Drop type for item
DROP TYPE IF EXISTS item_type;