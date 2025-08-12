# RuleStack POC — Weekend Build Task List

This document outlines the **full set of tasks** for building the RuleStack proof-of-concept in Go,
using `gorilla/mux` for routing, `sqlx` for Postgres access, and Flyway for migrations.

---

## 1. Environment Setup
- [ ] Create Go module: `go mod init rulestack`
- [ ] Add dependencies:
  ```bash
  go get github.com/gorilla/mux
  go get github.com/jmoiron/sqlx
  go get github.com/lib/pq
  go get github.com/spf13/cobra
  go get github.com/pelletier/go-toml/v2
  go get github.com/bmatcuk/doublestar/v4
  ```
- [ ] Install Flyway locally (CLI) for DB migrations
- [ ] Provision Postgres (local and/or Render)
- [ ] Create `.env` file with:
  ```env
  DATABASE_URL=postgres://user:pass@localhost:5432/rulestack?sslmode=disable
  TOKEN_SALT=some_random_string
  STORAGE_PATH=./storage
  PORT=8080
  ```

---

## 2. Database Migrations (Flyway)
- [ ] Create `migrations/` folder
- [ ] Add `V1__init_schema.sql`:
  ```sql
  CREATE TABLE packages (
      id SERIAL PRIMARY KEY,
      scope TEXT,
      name TEXT NOT NULL,
      UNIQUE (scope, name)
  );

  CREATE TABLE package_versions (
      id SERIAL PRIMARY KEY,
      package_id INT NOT NULL REFERENCES packages(id),
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

  CREATE TABLE tokens (
      id SERIAL PRIMARY KEY,
      token_hash TEXT NOT NULL,
      created_at TIMESTAMPTZ DEFAULT now()
  );
  ```
- [ ] Run migrations:
  ```bash
  flyway -url=jdbc:postgresql://localhost:5432/rulestack -user=postgres -password=secret migrate
  ```

---

## 3. API Service
**Location:** `cmd/api/main.go`

- [ ] Load config from env (`internal/config/config.go`)
- [ ] Connect to Postgres with `sqlx`
- [ ] Ensure `storage/` folder exists
- [ ] Set up `mux.Router`
- [ ] Middleware:
  - Logging
  - Auth check (Bearer token against `tokens` table)
- [ ] Routes:
  - `POST /v1/packages` – publish package
  - `GET /v1/packages` – search
  - `GET /v1/packages/{scope}/{name}` – package details
  - `GET /v1/packages/{scope}/{name}/versions/{version}` – version details
  - `GET /v1/blobs/{sha256}` – download blob
- [ ] Handlers:
  - **Publish**: parse manifest, save blob (hash, size, path), insert metadata
  - **Search**: basic SQL LIKE on name/description/tags
  - **Details**: lookup versions for a package
  - **Blob download**: stream file from `storage/`

---

## 4. CLI Tool
**Location:** `cmd/cli/main.go`

- [ ] Root command (`rfh`) with persistent flags:
  - `--registry`
  - `--token`
  - `-v` for verbose
- [ ] Config management (`internal/config/config.go`)
  - Stored at `~/.rfh/config.toml`
  - Commands:
    - `rfh registry add <name> <url> [token]`
    - `rfh registry use <name>`
    - `rfh registry list`
- [ ] Commands:
  1. **init** — create sample `rulestack.json`
  2. **pack** — tar.gz files from manifest
  3. **publish** — POST manifest + archive to registry
  4. **search** — GET from registry
  5. **add** — download and save ruleset (update `rfh.lock`)
  6. **apply** — extract ruleset into editor-specific path
- [ ] Internal helpers:
  - `internal/client/http.go` — API client (search, publish, etc.)
  - `internal/manifest/manifest.go` — load/validate manifest
  - `internal/pkg/archive.go` — pack/unpack tar.gz
  - `internal/pkg/lockfile.go` — manage `rfh.lock`
  - `internal/pkg/paths.go` — platform-safe install paths

---

## 5. Minimal End-to-End Flow (MVP)
1. Start API server locally:
   ```bash
   go run ./cmd/api
   ```
2. Add registry in CLI:
   ```bash
   rfh registry add default http://localhost:8080 my-token
   ```
3. Init a package:
   ```bash
   rfh init
   mkdir rules
   echo "Example rule" > rules/rule1.md
   ```
4. Pack and publish:
   ```bash
   rfh pack
   rfh publish
   ```
5. Search and install:
   ```bash
   rfh search "example"
   rfh add @acme/example@0.1.0
   rfh apply @acme/example --target cursor
   ```

---

## 6. Stretch Goals (Post-MVP)
- [ ] Dist-tags (`latest`, `beta`)
- [ ] Basic rule linting on publish
- [ ] Multiple registry support with per-scope mapping
- [ ] Private blob storage (S3)
- [ ] Signature verification (cosign)
- [ ] Transparency log for publishes
