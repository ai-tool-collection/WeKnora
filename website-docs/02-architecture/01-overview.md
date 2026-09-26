# Architecture

The Go API manages tenants, knowledge bases, ingestion, retrieval, agents, and integrations. The Vue frontend calls that API. PostgreSQL stores application data and can hold vectors with pgvector; Redis supports background work. Optional search and vector backends are selected through deployment configuration.

Documents pass through parsing, chunking, embedding, and indexing. At query time, retrieval selects relevant chunks; the configured model produces an answer with citations. Operators choose model providers and storage backends; no external model endpoint is required by the code itself.
