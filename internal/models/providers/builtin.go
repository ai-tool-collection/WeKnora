package providers

import "github.com/ai-tool-collection/WeKnora/internal/models/internal/configcopy"

// Builtins returns independent provider definitions without registering global state.
func Builtins() []*Definition {
	return configcopy.Clone([]*Definition{
		newAnthropicProvider(),
		newAzureOpenaiProvider(),
		newCohereProvider(),
		newGeminiProvider(),
		newGenericProvider(),
		newGpustackProvider(),
		newHuggingFaceTEIProvider(),
		newJinaProvider(),
		newLitellmProvider(),
		newNovitaProvider(),
		newNvidiaProvider(),
		newOpenaiProvider(),
		newOpenrouterProvider(),
		newPineconeProvider(),
		newRequestyProvider(),
	})
}
