# Milvus BM25 migration

This utility copies dense vectors from an older Milvus collection into a collection with language-aware BM25 fields. Keep the source collection and back up Milvus before changing the active collection prefix.

```sh
go run ./cmd/milvus-migrate --address 127.0.0.1:19530 --source weknora_embeddings --target knowledge_hub_embeddings
```

Use the same metric as the source index. Set `MILVUS_COLLECTION` to the target prefix only after verifying retrieval. The migration does not delete the old collection.
