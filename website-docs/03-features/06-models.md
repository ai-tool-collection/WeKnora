# Model connections

Configure chat, embedding, reranking, vision, and speech models in Settings → Models. Available integrations include OpenAI, Azure OpenAI, Anthropic, Gemini, OpenRouter, Ollama, and OpenAI-compatible endpoints. Choose the endpoint and credentials appropriate for your deployment.

Test the connection before assigning a model. Changing an embedding model usually requires rebuilding affected indexes because dimensions and semantic spaces can differ.

## Compatibility JSON {#compat-json}

OpenAI-compatible gateways can use compatibility settings to describe supported API behavior. Check the model editor for the fields accepted by the current version. Use a dedicated gateway URL and credentials for each deployment.
