-- Create packages table for storing package metadata
CREATE TABLE rulestack.packages (
    id SERIAL PRIMARY KEY,
    scope TEXT,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (scope, name)
);

-- Create package_versions table for storing version-specific data
CREATE TABLE rulestack.package_versions (
    id SERIAL PRIMARY KEY,
    package_id INT NOT NULL REFERENCES rulestack.packages(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    description TEXT,
    targets TEXT[],
    tags TEXT[],
    sha256 TEXT,
    size_bytes INT,
    blob_path TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (package_id, version)
);

-- Create tokens table for API authentication
CREATE TABLE rulestack.tokens (
    id SERIAL PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX idx_packages_scope_name ON rulestack.packages(scope, name);
CREATE INDEX idx_package_versions_package_id ON rulestack.package_versions(package_id);
CREATE INDEX idx_package_versions_version ON rulestack.package_versions(version);
CREATE INDEX idx_tokens_hash ON rulestack.tokens(token_hash);