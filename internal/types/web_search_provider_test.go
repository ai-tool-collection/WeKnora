package types

import "testing"

func TestGetWebSearchProviderTypesIncludesExa(t *testing.T) {
	var exa *WebSearchProviderTypeInfo
	providerTypes := GetWebSearchProviderTypes()
	for i := range providerTypes {
		if providerTypes[i].ID == string(WebSearchProviderTypeExa) {
			exa = &providerTypes[i]
			break
		}
	}
	if exa == nil {
		t.Fatal("Exa provider metadata is missing")
	}
	if !exa.RequiresAPIKey || !exa.SupportsProxy {
		t.Fatalf("unexpected Exa capability metadata: %+v", exa)
	}
	if len(exa.ConfigFields) != 1 {
		t.Fatalf("len(ConfigFields) = %d, want 1", len(exa.ConfigFields))
	}
	field := exa.ConfigFields[0]
	if field.Key != "include_text" || field.Type != "select" || field.Default != "false" {
		t.Fatalf("unexpected Exa config metadata: %+v", field)
	}
	if len(field.Options) != 2 || field.Options[0].Value != "true" || field.Options[1].Value != "false" {
		t.Fatalf("unexpected Exa config options: %+v", field.Options)
	}
}
func TestGetWebSearchProviderTypesIncludesSerply(t *testing.T) {
	var serply *WebSearchProviderTypeInfo
	providerTypes := GetWebSearchProviderTypes()
	for i := range providerTypes {
		if providerTypes[i].ID == string(WebSearchProviderTypeSerply) {
			serply = &providerTypes[i]
			break
		}
	}
	if serply == nil {
		t.Fatal("Serply provider type not found")
	}
	if !serply.RequiresAPIKey || !serply.SupportsProxy || serply.RequiresEngineID || serply.RequiresBaseURL {
		t.Fatalf("unexpected Serply metadata: %+v", serply)
	}
}
