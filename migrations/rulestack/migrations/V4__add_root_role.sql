-- V4__add_root_role.sql
-- Add root role to user_role enum

-- First, check if 'root' is not already in the enum and add it
DO $$
BEGIN
    -- Check if root is not in the enum
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_enum 
        WHERE enumtypid = 'rulestack.user_role'::regtype 
        AND enumlabel = 'root'
    ) THEN
        -- Add 'root' to the user_role enum
        ALTER TYPE rulestack.user_role ADD VALUE 'root';
    END IF;
END $$;

-- We need to commit the enum change before we can use it
-- This is handled by Flyway running each migration in its own transaction

-- Add comment to document the role hierarchy
COMMENT ON TYPE rulestack.user_role IS 'User roles: user (read-only), publisher (can publish packages), admin (user management), root (superuser with full access - only one allowed)';