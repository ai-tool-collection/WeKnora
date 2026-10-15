# Development

The API is written in Go and the frontend in Vue and TypeScript. Run Go tests with `go test ./...` in a configured Go environment. In `frontend/`, use `npm ci`, `npm run type-check`, and `npm run build`. The documentation site is built from `website-docs/` with `npm ci` and `npm run build`.

## Frontend dev proxy

The Vite dev server listens on `5173` and proxies `/api`, `/files`, and `/mcp/` to the backend. The target is the first non-empty value of `VITE_DEV_PROXY_TARGET`, `FRONTEND_BACKEND_URL`, or `http://localhost:8080`.

`make dev-frontend` loads the root `.env` and then `.env.local`, so a remote backend can be set in `.env.local`:

```bash
VITE_DEV_PROXY_TARGET=https://kb.example.com
```

When you run npm directly in `frontend/`, pass the variable explicitly, for example `VITE_DEV_PROXY_TARGET=https://kb.example.com npm run dev`. `npm run preview` (port `4173`) serves the production build with the same proxy rules. Restart the server after you change the target.
