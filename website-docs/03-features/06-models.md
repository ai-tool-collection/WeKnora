# Model connections

Configure chat, embedding, reranking, vision, and speech models in Settings → Models. Available integrations include OpenAI, Azure OpenAI, Anthropic, Gemini, OpenRouter, Ollama, and OpenAI-compatible endpoints. Choose the endpoint and credentials appropriate for your deployment.

Test the connection before assigning a model. Changing an embedding model usually requires rebuilding affected indexes because dimensions and semantic spaces can differ.

## Compatibility JSON {#compat-json}

OpenAI-compatible gateways can use compatibility settings to describe supported API behavior. Check the model editor for the fields accepted by the current version. Use a dedicated gateway URL and credentials for each deployment.

Rerank and embedding limits behave differently: `max_documents` and `max_request_chars` split a request into batches, while `max_query_chars` and `max_document_chars` fail the call when exceeded. Image-related fields (`image_format`, `image_field`, `image_prompt`, `text_as_messages`, `max_image_batch_size`, `max_image_bytes`, `image_mime_types`) apply only to models with image input.

## Rerank providers

Besides OpenAI-compatible and Jina-style endpoints, rerank models can use Cohere (v2 rerank API, text only), Hugging Face Text Embeddings Inference (TEI), and Pinecone Inference. Cohere requests split candidate sets larger than 1000 documents.

## Image input

Enable **Image input** on an embedding model only when it encodes images into the same vector space as text. Knowledge bases with image vector retrieval enabled then store one image vector per image after the image description is generated. Rerank models whose capabilities include image input also receive image documents.
