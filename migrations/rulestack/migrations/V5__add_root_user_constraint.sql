-- V5__add_root_user_constraint.sql
-- Add constraint to ensure only one root user exists

-- Add a constraint to ensure only one root user exists
-- First drop the index if it exists (for idempotency)
DROP INDEX IF EXISTS rulestack.only_one_root_user;

-- Add the unique partial index to enforce only one root user
CREATE UNIQUE INDEX only_one_root_user ON rulestack.users (role) WHERE role = 'root';