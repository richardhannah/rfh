-- Remove scope support from packages table
-- This migration removes the scope column and associated constraints/indexes

-- Drop the unique constraint that includes scope
ALTER TABLE rulestack.packages DROP CONSTRAINT packages_scope_name_key;

-- Drop the index on scope and name
DROP INDEX IF EXISTS rulestack.idx_packages_scope_name;

-- Drop the scope column
ALTER TABLE rulestack.packages DROP COLUMN IF EXISTS scope;

-- Add a new unique constraint on just the name
ALTER TABLE rulestack.packages ADD CONSTRAINT packages_name_unique UNIQUE (name);

-- Create a new index on just the name for performance
CREATE INDEX idx_packages_name ON rulestack.packages(name);