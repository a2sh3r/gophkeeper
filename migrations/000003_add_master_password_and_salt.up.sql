-- Add master password and salt columns to users table
ALTER TABLE users 
ADD COLUMN master_password VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN salt VARCHAR(255) NOT NULL DEFAULT '';

-- Update existing users with empty master password and salt
-- This will require users to re-register or we can set a default
UPDATE users SET 
    master_password = '',
    salt = ''
WHERE master_password = '' OR salt = '';

-- Make columns NOT NULL after setting defaults
ALTER TABLE users 
ALTER COLUMN master_password SET NOT NULL,
ALTER COLUMN salt SET NOT NULL;
