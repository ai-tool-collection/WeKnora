# Knowledge Hub

Knowledge Hub is a self-hosted knowledge base and agent application. It ingests documents, indexes their contents, and answers questions with source citations. The Go API and Vue frontend can run together with Docker Compose.

## Quickstart

1. Install Docker and Docker Compose.
2. Copy `.env.example` to `.env` and review passwords, ports, and outbound endpoints.
3. Start the stack:

   ```sh
   docker compose up -d --build
   ```

4. Open the frontend at `http://localhost` or the `FRONTEND_PORT` set in `.env`. Create the first administrator account and follow the setup wizard to connect a chat model and an embedding model.
5. Create a knowledge base, upload a document, and ask a question.

See the [quickstart](website-docs/01-getting-started/03-quickstart.md) and [troubleshooting guide](website-docs/01-getting-started/05-troubleshooting.md) for details.

## Components

- **Knowledge bases:** document ingestion, parsing, chunking, hybrid retrieval, citations, and optional wiki and knowledge graph features.
- **Agents:** model-backed chat, tools, MCP connections, and workspace skills.
- **Model connections:** OpenAI, Azure OpenAI, Anthropic, Gemini, OpenRouter, Ollama, and OpenAI-compatible endpoints, among others. No external model endpoint is hardcoded as a requirement.
- **Storage:** local files, MinIO, and S3-compatible object storage. Vector backends include pgvector, Elasticsearch, OpenSearch, Qdrant, Weaviate, and the self-hosted open-source Milvus and Apache Doris options.
- **Messaging:** Slack, Telegram, and Mattermost integrations.

Configure only the external services you use. Keep API keys on the server, use HTTPS for public deployments, and restrict outbound traffic according to your environment.

## Development

The API is in `internal/` and `cmd/`; the frontend is in `frontend/`; deployment files are in `docker/`, `docker-compose.yml`, and `helm/`. The documentation site is in `website-docs/`.

```sh
go test ./...
cd frontend && npm ci && npm run type-check && npm run build
```

The Go module path remains `github.com/ai-tool-collection/WeKnora` for import compatibility. Some environment variables and API identifiers retain their existing names for the same reason.

Maintainers importing upstream changes should follow the [China service removal and upstream update manual](docs/CHINA_SERVICE_REMOVAL.md).

## Documentation and license

Read the [documentation index](website-docs/index.md), [model guide](website-docs/03-features/06-models.md), and [API overview](website-docs/04-api/01-api-overview.md). This project is licensed under [MIT](LICENSE). Upstream attribution and license notices remain in their original legal files.
