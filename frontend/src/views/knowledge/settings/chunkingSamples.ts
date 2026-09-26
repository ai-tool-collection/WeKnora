// Representative English documents for previewing chunk boundaries.
export interface ChunkingSample {
  id: string
  labelKey: string
  text: string
}

const MARKDOWN_SAMPLE = `# Knowledge Hub architecture

Knowledge Hub indexes documents and answers questions with retrieved context. Teams can import files, web pages, Notion pages, and repositories into a knowledge base.

## Ingestion

An upload moves through parsing, text cleanup, chunking, embedding, and indexing. The parser extracts text and metadata. Chunking preserves headings and paragraph boundaries so that retrieved passages remain understandable.

## Retrieval

Queries can use semantic and keyword search. The results are merged, reranked, and sent to the selected language model with source references. Operators can inspect the returned passages when an answer needs review.

## Deployment

Run PostgreSQL, Redis, Qdrant, and the application with Docker Compose. Set model credentials through the application settings, then create a knowledge base and upload a document.

### Operations

Back up PostgreSQL and the object store regularly. Review service logs and tracing data when parsing or retrieval is slow.`

const FAQ_SAMPLE = `# Deployment FAQ

Q: How do I create a knowledge base?
A: Open Knowledge Bases, choose Create, enter a name, and select the desired embedding model.

Q: Why is my document still processing?
A: Check the parser service and application logs. Large files, encrypted PDFs, and unavailable model endpoints can delay parsing.

Q: Why are search results empty?
A: Confirm that indexing has completed and the vector store is reachable. Try a query that uses terms from the uploaded document.

Q: Can I change the chunk size?
A: Yes. Open the knowledge base settings and adjust the chunking configuration. New uploads use the updated setting.`

const CHAPTER_SAMPLE = `Chapter 1: Data lifecycle

1.1 Collection
Files and connected sources enter an ingestion queue. Each record has an origin, an update time, and a stable identifier.

1.2 Processing
The parser extracts content. A chunker divides long documents along structural boundaries. The embedding model turns each chunk into a vector.

Chapter 2: Search quality

2.1 Evaluation
Keep a small set of representative questions and expected source passages. Run them after changing models or chunking settings.

2.2 Maintenance
Reindex documents after changing the embedding model. Keep metadata filters aligned with access permissions.`

const PLAIN_SAMPLE = `Retrieval quality depends on chunk size and the embedding model. Very large chunks can mix unrelated topics. Very small chunks can lose the context needed to answer a question. Start with a moderate size and inspect real search results before tuning further.

Overlap helps when an answer crosses a chunk boundary. Too much overlap creates duplicate results and increases storage costs. A small overlap is usually enough for prose, while tables and code may need a different strategy.

When documents have headings, keep them with the paragraphs they introduce. Test searches with the questions users actually ask. Reranking can improve the order of candidate passages, but it adds latency and should be measured against the value it provides.`

export const CHUNKING_SAMPLES: ChunkingSample[] = [
  { id: 'markdown', labelKey: 'samples.markdown', text: MARKDOWN_SAMPLE },
  { id: 'faq', labelKey: 'samples.faq', text: FAQ_SAMPLE },
  { id: 'chapter', labelKey: 'samples.chapter', text: CHAPTER_SAMPLE },
  { id: 'plain', labelKey: 'samples.plain', text: PLAIN_SAMPLE },
]

export const DEFAULT_SAMPLE_ID = 'markdown'
