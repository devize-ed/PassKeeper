# Package migrations

Database schema migrations for PassKeeper.
Migrations are run automatically when the DB connection is created via `db.NewDB`.

## Migration files

- `000001_create_passkeeper_tables.up.sql` – creates users and items tables
- `000001_create_passkeeper_tables.down.sql` – drops tables
