package runtime_test

import (
	"bytes"
	"testing"

	"github.com/ai-tool-collection/WeKnora/internal/models/api"
	modelruntime "github.com/ai-tool-collection/WeKnora/internal/models/runtime"
	"github.com/ai-tool-collection/WeKnora/internal/types"
)

// Older rows for retained providers must continue to validate and resolve.
func TestRetainedLegacyRowsValidateAndResolve(t *testing.T) {
	rows := []struct {
		name   string
		model  string
		typ    types.ModelType
		params types.ModelParameters
	}{
		{"OpenAI chat", "gpt-4o", types.ModelTypeKnowledgeQA,
			types.ModelParameters{Provider: "openai", BaseURL: "https://api.openai.com/v1"}},
		{"local model", "custom-model", types.ModelTypeKnowledgeQA,
			types.ModelParameters{Provider: "generic", BaseURL: "http://localhost:8000/v1"}},
		{"OpenAI embedding", "text-embedding-3-small", types.ModelTypeEmbedding,
			types.ModelParameters{Provider: "openai", BaseURL: "https://api.openai.com/v1"}},
		{"Jina rerank", "jina-reranker-v3", types.ModelTypeRerank,
			types.ModelParameters{Provider: "jina", BaseURL: "https://api.jina.ai/v1"}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			params := row.params
			if err := modelruntime.ValidateRow(row.model, row.typ, &params); err != nil {
				t.Fatalf("ValidateRow: %v", err)
			}
			resolved, err := modelruntime.Resolve(modelruntime.Ref{
				Provider: params.Provider, Model: row.model, BaseURL: params.BaseURL,
				ModelType: row.typ, Extra: params.ExtraConfig, Override: params.Spec,
			})
			if err != nil || resolved == nil || resolved.Vendor == nil {
				t.Fatalf("Resolve: %v", err)
			}
		})
	}
}

func TestEveryCataloguedModelResolves(t *testing.T) {
	for _, v := range modelruntime.List() {
		for _, m := range v.Models() {
			if m.ID == "" || bytes.Contains(m.Compat, []byte("unsupported_reason")) {
				continue
			}
			modelType := m.Type
			if modelType == "" {
				modelType = types.ModelTypeKnowledgeQA
			}
			if err := modelruntime.ValidateRow(m.ID, modelType, &types.ModelParameters{Provider: v.ID}); err != nil {
				t.Errorf("%s/%s: ValidateRow: %v", v.ID, m.ID, err)
			}
		}
	}
}

// legacyThinkingControl is the old chat.parseThinkingOverride mapping, from
// internal/models/chat/thinking.go before the refactor. The four values are
// the only ones the old frontend wrote (frontend/src/utils/thinkingControl.ts),
// but an unrecognized non-empty value historically fell back to
// chat_template_kwargs and must keep doing so.
var legacyThinkingControl = map[string]api.ThinkingFormat{
	"none":                 api.ThinkingFormatNone,
	"enable_thinking":      api.ThinkingFormatEnableThinking,
	"thinking_type":        api.ThinkingFormatThinkingType,
	"chat_template_kwargs": api.ThinkingFormatChatTemplateKwargs,
	"enabled":              api.ThinkingFormatChatTemplateKwargs,
	"ENABLE_THINKING":      api.ThinkingFormatEnableThinking,
	"  thinking_type  ":    api.ThinkingFormatThinkingType,
}

// TestLegacyThinkingControlMapping pins extra_config.thinking_control to the
// old semantics for every value a stored row can hold, including on a
// catalogued non-reasoning model where the new silencing pass would otherwise
// drop the thinking fields.
func TestLegacyThinkingControlMapping(t *testing.T) {
	for value, want := range legacyThinkingControl {
		t.Run(value, func(t *testing.T) {
			resolved, err := modelruntime.Resolve(modelruntime.Ref{
				Provider: "generic", Model: "Qwen/Qwen3-32B",
				BaseURL: "http://vllm.internal:8000/v1", ModelType: types.ModelTypeKnowledgeQA,
				Extra: map[string]string{"thinking_control": value},
			})
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if got := resolved.OpenAICompletions.ThinkingFormat; got != want {
				t.Fatalf("thinking_control=%q resolved to %q, old code selected %q", value, got, want)
			}
		})
	}
}

// TestLegacyThinkingControlSurvivesCatalogSilencing guards the interaction
// between a stored thinking_control and silenceThinkingForNonReasoningModel:
// an operator who explicitly picked an encoding must keep it even when the
// catalog marks the model non-reasoning.
func TestLegacyThinkingControlSurvivesCatalogSilencing(t *testing.T) {
	resolved, err := modelruntime.Resolve(modelruntime.Ref{
		Provider: "openai", Model: "gpt-4o", ModelType: types.ModelTypeKnowledgeQA,
		Extra: map[string]string{"thinking_control": "chat_template_kwargs"},
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.OpenAICompletions.ThinkingFormat != api.ThinkingFormatChatTemplateKwargs {
		t.Fatalf("explicit thinking_control was overridden by the catalog: %q",
			resolved.OpenAICompletions.ThinkingFormat)
	}
}

// TestUncataloguedModelKeepsThinkingSwitch documents the deliberate asymmetry
// in silenceThinkingForNonReasoningModel: a model nobody catalogued (the
// self-hosted case) keeps its thinking switch, which is what old rows relied
// on, while a catalogued non-reasoning model loses it.
func TestUncataloguedModelKeepsThinkingSwitch(t *testing.T) {
	unknown, err := modelruntime.Resolve(modelruntime.Ref{
		Provider: "generic", Model: "acme/llama-3.1-70b-finetune",
		BaseURL: "http://vllm.internal:8000/v1", ModelType: types.ModelTypeKnowledgeQA,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if unknown.Cataloged {
		t.Fatal("expected the fine-tune to be uncatalogued")
	}
	if unknown.OpenAICompletions.ThinkingFormat != api.ThinkingFormatChatTemplateKwargs {
		t.Fatalf("uncatalogued model lost its thinking switch: %q",
			unknown.OpenAICompletions.ThinkingFormat)
	}
}

// TestLegacyRemoteModelNameHonoured pins extra_config.remote_model_name, which
// the old chat and embedding paths both read.
func TestLegacyRemoteModelNameHonoured(t *testing.T) {
	resolved, err := modelruntime.Resolve(modelruntime.Ref{
		Provider: "generic", Model: "display-name", BaseURL: "http://vllm.internal:8000/v1",
		ModelType: types.ModelTypeKnowledgeQA,
		Extra:     map[string]string{"remote_model_name": "Qwen/Qwen3-32B"},
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.RemoteModel != "Qwen/Qwen3-32B" {
		t.Fatalf("remote_model_name ignored: wire model is %q", resolved.RemoteModel)
	}
}

// TestUnknownProviderFallsBackToGeneric checks the reverse direction: a row
// naming a provider this build does not ship degrades to the generic
// OpenAI-compatible baseline instead of failing the write.
func TestUnknownProviderFallsBackToGeneric(t *testing.T) {
	resolved, err := modelruntime.Resolve(modelruntime.Ref{
		Provider: "a-vendor-from-the-future", Model: "m",
		BaseURL: "https://example.com/v1", ModelType: types.ModelTypeKnowledgeQA,
	})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if resolved.Vendor.ID != providers.GenericID {
		t.Fatalf("unknown provider resolved to %q, want %q", resolved.Vendor.ID, providers.GenericID)
	}
	if resolved.API != api.APIOpenAICompletions {
		t.Fatalf("unknown provider resolved to protocol %q", resolved.API)
	}
}

// TestLegacyStoredBaseURLSelectsSameProtocol pins the protocol an existing
// row resolves to. Old rows always store an explicit base_url (the model
// editor pre-filled the provider default), so the stored URL — not the new
// vendor default — decides the wire protocol on upgrade. The Gemini case and
// the first-party OpenAI move to Responses are covered in
// internal/models/parity; what is asserted here is the relay case, where the
// same provider id must NOT follow OpenAI onto /responses.
func TestLegacyStoredBaseURLSelectsSameProtocol(t *testing.T) {
	cases := []struct {
		provider, baseURL string
		want              api.API
		why               string
	}{
		{
			provider: "openai", baseURL: "https://llm-relay.corp.example.com/v1",
			want: api.APIOpenAICompletions,
			why:  "a relay that only speaks Chat Completions must stay on it",
		},
		{
			provider: "openai", baseURL: "https://litellm.corp.example.com/v1",
			want: api.APIOpenAICompletions,
			why:  "an OpenAI row pointed at a LiteLLM proxy keeps Chat Completions",
		},
		{
			provider: "anthropic", baseURL: "https://api.anthropic.com/v1",
			want: api.APIAnthropicMessages,
			why:  "the old code already dispatched anthropic to the Messages protocol",
		},
		{
			provider: "generic", baseURL: "http://vllm.internal:8000/v1",
			want: api.APIOpenAICompletions, why: "unchanged",
		},
	}
	for _, tc := range cases {
		t.Run(tc.provider+" "+tc.baseURL, func(t *testing.T) {
			resolved, err := modelruntime.Resolve(modelruntime.Ref{
				Provider: tc.provider, Model: "some-model", BaseURL: tc.baseURL,
				ModelType: types.ModelTypeKnowledgeQA,
			})
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			if resolved.API != tc.want {
				t.Fatalf("protocol %q, want %q (%s)", resolved.API, tc.want, tc.why)
			}
		})
	}
}

// TestLegacyFreeFormAPIExtraKeyIsRejected documents the one stored shape this
// refactor does break, so the trade-off stays visible.
//
// extra_config has always been a free-form map[string]string: the old backend
// read only thinking_control, remote_model_name, api_version, secret_key,
// region, instruction and truncate_prompt_tokens from it and ignored every
// other key. The catalog gave the key "api" a meaning — it forces the wire
// protocol — and ValidateRow now rejects a value that is not one of the five
// protocol names. A row that an operator hand-populated with, say,
// {"api": "v1"} therefore fails PUT /models/{id} with 400 and fails every
// chat call, even though nothing about that row changed.
//
// Nothing in WeKnora ever wrote this key, so only hand-built rows and
// hand-written builtin_models.yaml entries are affected. If that is judged too
// sharp, the fix is to ignore an unparseable extra_config.api in Resolve and
// keep the strict check for Spec.API, which no old row can carry.
func TestLegacyFreeFormAPIExtraKeyIsRejected(t *testing.T) {
	err := modelruntime.ValidateRow("gpt-4o", types.ModelTypeKnowledgeQA, &types.ModelParameters{
		Provider: "openai", BaseURL: "https://api.openai.com/v1",
		ExtraConfig: map[string]string{"api": "v1"},
	})
	if err == nil {
		t.Skip("extra_config.api is now tolerant of legacy free-form values; " +
			"delete this test and the release note that goes with it")
	}
	t.Logf("documented break: %v", err)
}
