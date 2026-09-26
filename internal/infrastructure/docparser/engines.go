package docparser

import (
	"context"
	"strings"

	"github.com/ai-tool-collection/WeKnora/internal/infrastructure/docparser/anydoc"
	"github.com/ai-tool-collection/WeKnora/internal/types"
	"github.com/ai-tool-collection/WeKnora/internal/types/interfaces"
)

// Engine names, as stored in knowledge base parser rules and shown in the UI.
const (
	// BuiltinEngineName is the DocReader (Python) parser suite.
	BuiltinEngineName = "builtin"
	// SimpleEngineName is Go-native handling of text formats and images.
	SimpleEngineName = "simple"
	// AnydocEngineName is the in-process anydoc office-document converter.
	AnydocEngineName = "anydoc"
)

func init() {
	types.SetPreferParserEngine(preferAnydocWhenAvailable)
	RegisterEngine(&builtinEngine{})
	RegisterEngine(&simpleEngine{})
	RegisterEngine(&anydocEngine{})
}

// preferAnydocWhenAvailable is the type-level default override: when the
// anydoc binding is linked and converts this file type, use it instead of
// builtin or the markitdown fallback. Simple formats stay on the Go reader.
//
// PDF is excluded. anydoc has no document model for it, so Convert force-
// disables asset extraction, and AnydocReader only falls back to builtin
// when the whole file yields no text at all. A born-digital PDF that yields
// a few characters is therefore reported as parsed while its figures,
// tables and layout are gone. The builtin parser classifies pages
// individually and routes scanned ones through OCR/VLM, so it is the better
// default for the one type anydoc cannot model.
func preferAnydocWhenAvailable(fileType string) string {
	ft := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(fileType), "."))
	if IsSimpleFormat(ft) || ft == "pdf" {
		return ""
	}
	if anydoc.Available() && anydoc.Supports(ft, "") {
		return AnydocEngineName
	}
	return ""
}

// ---------------------------------------------------------------------------
// builtin — DocReader-backed parser for complex document formats.
// ---------------------------------------------------------------------------

type builtinEngine struct{}

func (e *builtinEngine) Name() string { return BuiltinEngineName }

func (e *builtinEngine) Description() string { return "DocReader built-in parser engine" }

func (e *builtinEngine) FileTypes(_ bool) []string {
	return []string{
		"docx", "doc", "pdf", "md", "markdown", "xlsx", "xls",
		"pptx", "ppt", "epub",
		"html", "htm", "mhtml", "xmind",
		"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp",
		"mp3", "wav", "m4a", "flac", "ogg",
	}
}

func (e *builtinEngine) CheckAvailable(docreaderConnected bool, _ map[string]string) (bool, string) {
	if docreaderConnected {
		return true, ""
	}
	return false, "DocReader service not connected"
}

// NewReader returns the docreader client. Selecting "builtin" explicitly means
// the docreader is wanted even for formats the simple reader could handle.
func (e *builtinEngine) NewReader(_ context.Context, deps ReaderDeps) (interfaces.DocReader, error) {
	return remoteReader(deps)
}

// ---------------------------------------------------------------------------
// simple — Go handles md/txt/csv/json natively, no external service needed.
// Distinct from docreader's "builtin", which uses Python libraries for complex
// formats (docx, pdf, ...).
// ---------------------------------------------------------------------------

type simpleEngine struct{}

func (e *simpleEngine) Name() string { return SimpleEngineName }

func (e *simpleEngine) Description() string {
	return "Simple format & image parsing (no external service required)"
}

func (e *simpleEngine) FileTypes(_ bool) []string {
	return []string{
		"md", "markdown", "txt", "csv", "json",
		"jpg", "jpeg", "png", "gif", "bmp", "tiff", "webp",
		"mp3", "wav", "m4a", "flac", "ogg",
	}
}

func (e *simpleEngine) CheckAvailable(_ bool, _ map[string]string) (bool, string) {
	return true, ""
}

func (e *simpleEngine) NewReader(_ context.Context, _ ReaderDeps) (interfaces.DocReader, error) {
	return &SimpleFormatReader{}, nil
}

// ---------------------------------------------------------------------------
// anydoc — office documents converted in this process, no external service.
// Only present in builds that link the converter (see the anydoc package).
// ---------------------------------------------------------------------------

type anydocEngine struct{}

func (e *anydocEngine) Name() string { return AnydocEngineName }

func (e *anydocEngine) Description() string {
	return "anydoc in-process office document converter (no external service required)"
}

func (e *anydocEngine) FileTypes(_ bool) []string { return anydoc.SupportedFileTypes() }

func (e *anydocEngine) CheckAvailable(_ bool, _ map[string]string) (bool, string) {
	if anydoc.Available() {
		return true, ""
	}
	return false, anydoc.UnavailableReason()
}

func (e *anydocEngine) NewReader(_ context.Context, deps ReaderDeps) (interfaces.DocReader, error) {
	if !anydoc.Available() {
		return nil, errEngineUnavailable(AnydocEngineName, anydoc.UnavailableReason())
	}
	return NewAnydocReader(deps.Overrides, deps.Remote), nil
}
