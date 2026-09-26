# Troubleshooting

Check `docker compose ps` and `docker compose logs` for the failing service. Validate model endpoints from the application host and confirm that the configured embedding dimension matches existing indexes.

## Database migrations

The application runs database migrations during startup. If startup stops at a migration, back up the database, inspect the failing migration in the logs, and correct the underlying schema or permission issue before restarting. Avoid deleting the database to work around migration errors.
