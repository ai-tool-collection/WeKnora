# Quickstart

## Requirements

Install Docker and Docker Compose. Copy `.env.example` to `.env` at the repository root, then set passwords and any model credentials required by your deployment.

## Run

```sh
docker compose up -d --build
```

Open the frontend port configured in `docker-compose.yml`, create the administrator account, and use the initialization wizard to connect a chat model and an embedding model. Create a knowledge base, upload a document, and ask a question to verify citations.

For a local model, configure an Ollama or OpenAI-compatible endpoint. Keep credentials on the server. See [model connections](../03-features/06-models.md).
