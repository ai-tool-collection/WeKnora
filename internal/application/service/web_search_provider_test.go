package service

import (
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/types"
)

func TestValidateProviderParametersExa(t *testing.T) {
	valid := types.WebSearchProviderParameters{APIKey: "exa-test"}
	if err := validateProviderParameters(types.WebSearchProviderTypeExa, valid); err != nil {
		t.Fatalf("valid Exa parameters rejected: %v", err)
	}
	if !isValidProviderType(types.WebSearchProviderTypeExa) {
		t.Fatal("Exa provider type is not accepted")
	}

	if err := validateProviderParameters(types.WebSearchProviderTypeExa, types.WebSearchProviderParameters{}); err == nil {
		t.Fatal("missing Exa API key was accepted")
	}
}

func TestValidateProviderParametersSerply(t *testing.T) {
	valid := types.WebSearchProviderParameters{APIKey: "serply-test"}
	if err := validateProviderParameters(types.WebSearchProviderTypeSerply, valid); err != nil {
		t.Fatalf("valid Serply parameters rejected: %v", err)
	}
	if !isValidProviderType(types.WebSearchProviderTypeSerply) {
		t.Fatal("Serply provider type is not accepted")
	}
	blank := types.WebSearchProviderParameters{APIKey: "   "}
	if err := validateProviderParameters(types.WebSearchProviderTypeSerply, blank); err == nil {
		t.Fatal("blank Serply API key was accepted")
	}
}
