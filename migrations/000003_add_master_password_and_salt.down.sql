-- Remove master password and salt columns from users table
ALTER TABLE users 
DROP COLUMN IF EXISTS master_password,
DROP COLUMN IF EXISTS salt;
