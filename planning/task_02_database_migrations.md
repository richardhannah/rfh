# Task 2: Database Migrations (1 hour)

## Objective
Create and run database migrations using Flyway to set up the RuleStack schema.

## Prerequisites
- Task 1 completed (environment setup)
- Flyway CLI installed and accessible
- Database connection string available

## Checklist

### 1. Create Migration Directory (5 minutes)
- [ ] Ensure `migrations/` directory exists in project root
- [ ] Verify Flyway can access this directory

### 2. Create Initial Schema Migration (30 minutes)
Create `migrations/V1__init_schema.sql`:
```sql
-- Create packages table for storing package metadata
CREATE TABLE packages (
    id SERIAL PRIMARY KEY,
    scope TEXT,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (scope, name)
);

-- Create package_versions table for storing version-specific data
CREATE TABLE package_versions (
    id SERIAL PRIMARY KEY,
    package_id INT NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
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
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    name TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for better query performance
CREATE INDEX idx_packages_scope_name ON packages(scope, name);
CREATE INDEX idx_package_versions_package_id ON package_versions(package_id);
CREATE INDEX idx_package_versions_version ON package_versions(version);
CREATE INDEX idx_tokens_hash ON tokens(token_hash);
```

- [ ] Create the migration file with above content
- [ ] Review SQL syntax for any errors

### 3. Run Migration (10 minutes)
Test migration with your database:

**Local Postgres:**
```bash
flyway -url=jdbc:postgresql://localhost:5432/rulestack -user=postgres -password=yourpassword migrate
```

**Remote Postgres (adjust URL):**
```bash
flyway -url="jdbc:postgresql://your-host:5432/rulestack?sslmode=require" -user=your-user -password=your-password migrate
```

- [ ] Run flyway migrate command
- [ ] Verify success message: "Migration completed successfully"
- [ ] Check tables created: `psql -d rulestack -c "\dt"`

### 4. Verify Schema (10 minutes)
Check that all tables and indexes were created:
```sql
-- List all tables
\dt

-- Describe packages table
\d packages

-- Describe package_versions table  
\d package_versions

-- Describe tokens table
\d tokens

-- List indexes
\di
```
- [ ] All 3 tables exist (`packages`, `package_versions`, `tokens`)
- [ ] Foreign key constraints are in place
- [ ] Indexes are created
- [ ] UNIQUE constraints are active

### 5. Seed Test Data (Optional - 5 minutes)
Create a test token for development:
```sql
INSERT INTO tokens (token_hash, name) 
VALUES ('test_hash_12345', 'development_token');
```
- [ ] Insert test token (or skip if not needed immediately)

## Validation Commands
```bash
# Check migration status
flyway -url=jdbc:postgresql://localhost:5432/rulestack -user=postgres -password=yourpassword info

# Verify tables exist
psql -d rulestack -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"

# Check if we can insert/query
psql -d rulestack -c "INSERT INTO packages (name) VALUES ('test-package'); SELECT * FROM packages;"
```

## Troubleshooting
- **Connection issues**: Verify DATABASE_URL format and credentials
- **Permission denied**: Ensure database user has CREATE privileges  
- **Migration fails**: Check SQL syntax and table constraints
- **Flyway not found**: Verify Flyway is in PATH

## Acceptance Criteria
- [ ] Migration V1 runs successfully
- [ ] All 3 tables created with correct schema
- [ ] Foreign key relationships established
- [ ] Indexes created for performance
- [ ] Can insert and query test data
- [ ] Flyway tracks migration state correctly

## Time Estimate: ~60 minutes

## Next Task  
Task 3: Basic Configuration Module