-- Create user roles enum
CREATE TYPE rulestack.user_role AS ENUM ('user', 'publisher', 'admin');

-- Create users table
CREATE TABLE rulestack.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role rulestack.user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    last_login TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT true
);

-- Create user_sessions table for JWT tokens
CREATE TABLE rulestack.user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES rulestack.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    last_used TIMESTAMPTZ DEFAULT now(),
    user_agent TEXT,
    ip_address INET
);

-- Update tokens table to link with users (for backward compatibility)
ALTER TABLE rulestack.tokens ADD COLUMN user_id INT REFERENCES rulestack.users(id) ON DELETE CASCADE;
ALTER TABLE rulestack.tokens ADD COLUMN expires_at TIMESTAMPTZ;

-- Create indexes for better performance
CREATE INDEX idx_users_username ON rulestack.users(username);
CREATE INDEX idx_users_email ON rulestack.users(email);
CREATE INDEX idx_users_role ON rulestack.users(role);
CREATE INDEX idx_users_active ON rulestack.users(is_active);
CREATE INDEX idx_user_sessions_user_id ON rulestack.user_sessions(user_id);
CREATE INDEX idx_user_sessions_token_hash ON rulestack.user_sessions(token_hash);
CREATE INDEX idx_user_sessions_expires_at ON rulestack.user_sessions(expires_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION rulestack.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON rulestack.users
    FOR EACH ROW EXECUTE FUNCTION rulestack.update_updated_at_column();