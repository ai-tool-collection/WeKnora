package handler

import "github.com/ai-tool-collection/WeKnora/internal/storageallowlist"

func getSupportedStorageProviders() []string {
	return storageallowlist.Supported()
}

func getAllowedStorageProviders() map[string]bool {
	return storageallowlist.AllowedMap()
}

func isStorageProviderAllowed(provider string) bool {
	return storageallowlist.IsAllowed(provider)
}

func firstAllowedStorageProvider() string {
	return storageallowlist.FirstAllowed()
}
