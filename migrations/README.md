# Database Migrations

This directory contains database migration files structured for Flyway UI compatibility.

## Structure

```
migrations/
├── README.md                     # This file
└── rulestack/                   # Flyway UI project structure
    ├── flyway.conf             # Flyway configuration (Java properties format)
    ├── flyway.toml             # Flyway project configuration (TOML format)
    ├── flyway.user.toml        # User-specific Flyway settings
    ├── migrations/             # SQL migration files
    │   ├── V1__init_schema.sql
    │   └── V2__user_auth_system.sql
    └── schema-model/           # Flyway schema model (if used)
```

## Local Development

The Docker Compose setup uses Flyway for migrations:
- Flyway migration runner: `./migrations/rulestack/migrations:/flyway/sql`
- Database name: `rulestackdb` (matches production)
- Schema name: `rulestack` (all tables created in this schema)
- PostgreSQL creates empty database, Flyway creates schema and applies all migrations
- API connection uses `search_path=rulestack` to default to rulestack schema

## Production Deployment

For production deployment using Flyway UI:
1. The migrations folder structure is already compatible with Flyway UI expectations
2. Import the `migrations/rulestack` folder as a Flyway project
3. The project includes both configuration formats:
   - `flyway.conf`: Traditional Java properties format (for older Flyway UI versions)
   - `flyway.toml`: Modern TOML format with predefined environments (rulestack-local, rulestack-prod)
4. Configure your production database connection:
   - For `.conf` format: Update the `flyway.url`, `flyway.user`, and `flyway.password` properties
   - For `.toml` format: Update the `[environments.rulestack-prod]` section
5. Run migrations through the Flyway UI interface

## Adding New Migrations

1. Create new migration files in `migrations/rulestack/migrations/`
2. Follow Flyway naming convention: `V{version}__{description}.sql`
3. Test locally using `docker-compose up` to verify migrations work
4. Deploy to production using Flyway UI