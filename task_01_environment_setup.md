# Task 1: Environment Setup (1 hour)

## Objective
Set up the Go development environment and dependencies for RuleStack POC.

## Prerequisites
- Go 1.21+ installed
- Postgres database available (local or remote)
- Flyway CLI installed

## Checklist

### 1. Initialize Go Module (5 minutes)
- [ ] Create new directory: `mkdir rulestack && cd rulestack`
- [ ] Initialize Go module: `go mod init rulestack`
- [ ] Verify `go.mod` file created

### 2. Add Dependencies (10 minutes)
Add all required Go dependencies:
```bash
go get github.com/gorilla/mux
go get github.com/jmoiron/sqlx
go get github.com/lib/pq
go get github.com/spf13/cobra
go get github.com/pelletier/go-toml/v2
go get github.com/bmatcuk/doublestar/v4
```
- [ ] Run the commands above
- [ ] Verify all dependencies in `go.mod` and `go.sum`

### 3. Install Flyway CLI (15 minutes)
- [ ] Download Flyway CLI from https://flywaydb.org/download/
- [ ] Install or extract to PATH
- [ ] Test: `flyway -v` shows version

### 4. Database Setup (15 minutes)
**Local Postgres (if using):**
- [ ] Start local Postgres service
- [ ] Create database: `createdb rulestack`
- [ ] Test connection: `psql -d rulestack -c "SELECT 1;"`

**Or Render Postgres:**
- [ ] Create Postgres instance on Render
- [ ] Copy connection string from Render dashboard

### 5. Environment Configuration (10 minutes)
Create `.env` file in project root:
```env
DATABASE_URL=postgres://user:pass@localhost:5432/rulestack?sslmode=disable
TOKEN_SALT=your_random_salt_string_here
STORAGE_PATH=./storage
PORT=8080
```
- [ ] Create `.env` file with actual values
- [ ] Create `.gitignore` with `.env` entry
- [ ] Create `storage/` directory: `mkdir storage`

### 6. Project Structure (5 minutes)
Create basic directory structure:
```
rulestack/
├── cmd/
│   ├── api/
│   └── cli/
├── internal/
│   ├── api/
│   ├── config/
│   ├── db/
│   ├── storage/
│   ├── client/
│   ├── manifest/
│   └── pkg/
├── migrations/
├── storage/
├── .env
├── .gitignore
├── go.mod
└── go.sum
```
- [ ] Create all directories using `mkdir -p`

## Acceptance Criteria
- [ ] Go module initialized with all dependencies
- [ ] Flyway CLI accessible and working
- [ ] Database connection verified
- [ ] Environment file configured
- [ ] Project structure created
- [ ] Can run `go mod tidy` without errors

## Time Estimate: ~60 minutes

## Next Task
Task 2: Database Migrations