# Seafile data source

The `seafile` connector syncs selected folders and files from one self-hosted Seafile library into a knowledge base. It supports scheduled, incremental, and resumable sync, and can reconcile deletions. The compatibility baseline is Seafile Community Server 10.0.1 (`api2`).

## Credentials

Use an account with at least read access to the target library. The connector stores only that account's user API token; it does not handle passwords, SSO, or library-specific tokens. The connection test calls `/api2/auth/ping/`. Library and folder access is checked when the resource tree loads.

Seafile 10.0.1 hides the Web API token by default. An administrator can set `ENABLE_GET_AUTH_TOKEN_BY_SESSION = True` in `seahub_settings.py`, or you can request a token with `POST /api2/auth-token/`.

## Setup

1. Open the knowledge base settings, go to **Data sources**, and add **Seafile**.
2. Enter the Seafile URL (a path prefix such as `https://example.com/seafile` is allowed) and the API token. Test the connection.
3. Select folders or files from one unencrypted library. A selected folder syncs recursively. Only file types that the knowledge base can import are listed.
4. Choose the sync mode, schedule, conflict policy, and deletion sync. Run a first sync with a few files and check the logs.

A data source cannot switch to another library after it has synced. Create a new data source instead.

## Network

API requests go to the configured URL with `Authorization: Token`. File downloads use the fileserver URL that Seafile returns (usually `/seafhttp/…`) without the token. Both pass SSRF validation. For a private deployment or a separate fileserver domain, add the hosts to `SSRF_WHITELIST`; otherwise saving fails with `base_url SSRF validation failed` or the sync log shows `seafile_ssrf_blocked`.

## Sync behavior

- Files land under `<library name>/<path>`. The library name is fixed at the first sync.
- Each file is fingerprinted by object ID, modification time, and size. Unchanged files are not downloaded again; changed files replace the old entry.
- Renames and moves import the new path and treat the old path as deleted.
- `MAX_FILE_SIZE_MB` limits each file. If you raise it, also raise `DOCREADER_GRPC_MAX_FILE_SIZE_MB` and restart the services.
- Per-file errors are logged and retried on the next sync. Folder listing errors stop the run without deletions. Requests with 429, 502, 503, or 504 are retried up to 3 times.
- Checkpoints are saved during the run, so an interrupted sync resumes from the last checkpoint.

Sync log error codes: `seafile_permission_denied`, `seafile_not_found`, `seafile_file_too_large`, `seafile_empty_file`, `seafile_source_changed`, `seafile_invalid_response`, `seafile_ssrf_blocked`, and `seafile_fetch_failed`.
