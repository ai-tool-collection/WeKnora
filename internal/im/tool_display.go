package im

import (
	"encoding/json"
	"fmt"
	"strings"
)

// IMToolStep tracks one tool invocation for IM display (mirrors Web agent stream events).
type IMToolStep struct {
	ToolCallID string
	ToolName   string
	Pending    bool
	Success    bool
	Arguments  map[string]any
	Data       map[string]interface{}
	Output     string
}

// imLocalizedToolName returns English labels for IM tool progress.
func imLocalizedToolName(toolName string) string {
	if name, ok := imToolNameLabels[toolName]; ok {
		return name
	}
	if strings.HasPrefix(toolName, "mcp_") {
		return formatMCPToolName(toolName)
	}
	return toolName
}

var imToolNameLabels = map[string]string{
	"search_knowledge":        "Search knowledge",
	"knowledge_search":        "Search knowledge",
	"grep_chunks":             "Search keywords",
	"read_document":           "Read document",
	"list_documents":          "Browse documents",
	"web_search":              "Search web",
	"web_fetch":               "Fetch web page",
	"get_document_info":       "Get document details",
	"list_knowledge_chunks":   "View knowledge chunks",
	"get_related_documents":   "Find related documents",
	"get_document_content":    "Get document content",
	"wiki_search":             "Search wiki",
	"wiki_read_page":          "Read wiki page",
	"wiki_read_source_doc":    "Read source document",
	"todo_write":              "Manage plan",
	"knowledge_graph_extract": "Extract knowledge graph",
	"thinking":                "Think",
	"image_analysis":          "Analyze image",
	"query_understand":        "Understand question",
	"query_knowledge_graph":   "Query knowledge graph",
	"read_skill":              "Read skill",
	"execute_skill_script":    "Run skill script",
	"list_sandbox_files":      "List sandbox files",
	"read_sandbox_file":       "Read sandbox file",
	"write_sandbox_file":      "Write sandbox file",
	"edit_sandbox_file":       "Edit sandbox file",
	"shell_exec":              "Run sandbox command",
	"data_analysis":           "Analyze data",
	"data_schema":             "View data schema",
	"database_query":          "Query database",
}

func formatMCPToolName(rawName string) string {
	rest := strings.TrimPrefix(rawName, "mcp_")
	if rest == "" {
		return rawName
	}
	parts := strings.Split(rest, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func collectQueryStrings(value any) []string {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		if strings.HasPrefix(trimmed, "[") {
			var parsed []any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
				var out []string
				for _, item := range parsed {
					if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
						out = append(out, strings.TrimSpace(s))
					}
				}
				return out
			}
		}
		return []string{trimmed}
	case []any:
		var out []string
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	case []string:
		var out []string
		for _, s := range v {
			if strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func imGetQueryText(args any) string {
	if args == nil {
		return ""
	}
	parsed := args
	if s, ok := args.(string); ok {
		var obj map[string]any
		if err := json.Unmarshal([]byte(s), &obj); err != nil {
			return ""
		}
		parsed = obj
	}
	record, ok := parsed.(map[string]any)
	if !ok {
		return ""
	}
	seen := make(map[string]struct{})
	var queries []string
	for _, q := range collectQueryStrings(record["query"]) {
		if _, dup := seen[q]; dup {
			continue
		}
		seen[q] = struct{}{}
		queries = append(queries, q)
	}
	for _, q := range collectQueryStrings(record["queries"]) {
		if _, dup := seen[q]; dup {
			continue
		}
		seen[q] = struct{}{}
		queries = append(queries, q)
	}
	return strings.Join(queries, ", ")
}

func imGetWikiPageText(args any) string {
	if args == nil {
		return ""
	}
	parsed := args
	if s, ok := args.(string); ok {
		var obj map[string]any
		if err := json.Unmarshal([]byte(s), &obj); err != nil {
			return ""
		}
		parsed = obj
	}
	record, ok := parsed.(map[string]any)
	if !ok {
		return ""
	}
	seen := make(map[string]struct{})
	var slugs []string
	for _, slug := range collectQueryStrings(record["slug"]) {
		if _, dup := seen[slug]; dup {
			continue
		}
		seen[slug] = struct{}{}
		slugs = append(slugs, slug)
	}
	for _, slug := range collectQueryStrings(record["slugs"]) {
		if _, dup := seen[slug]; dup {
			continue
		}
		seen[slug] = struct{}{}
		slugs = append(slugs, slug)
	}
	return strings.Join(slugs, ", ")
}

func imGetGrepPatterns(args any) []string {
	if args == nil {
		return nil
	}
	record, ok := args.(map[string]any)
	if !ok {
		return nil
	}
	if queries := collectQueryStrings(record["queries"]); len(queries) > 0 {
		return queries
	}
	if patterns := collectQueryStrings(record["patterns"]); len(patterns) > 0 {
		return patterns
	}
	if q := imGetQueryText(record); q != "" {
		return []string{q}
	}
	if pattern, ok := record["pattern"].(string); ok && strings.TrimSpace(pattern) != "" {
		return []string{strings.TrimSpace(pattern)}
	}
	return nil
}

func imGetWebSearchQuery(step IMToolStep) string {
	if q := imGetQueryText(step.Arguments); q != "" {
		return q
	}
	return imGetQueryText(step.Data)
}

func imGetGrepPatternsFromStep(step IMToolStep) []string {
	if patterns := imGetGrepPatterns(step.Arguments); len(patterns) > 0 {
		return patterns
	}
	return imGetGrepPatterns(step.Data)
}

func imAppendQueryTitle(base, query string) string {
	if query == "" {
		return base
	}
	return fmt.Sprintf("%s: %s", base, query)
}

func imAppendPatternsTitle(base string, patterns []string) string {
	if len(patterns) == 0 {
		return base
	}
	display := patterns
	more := ""
	if len(patterns) > 2 {
		display = patterns[:2]
		more = fmt.Sprintf(" +%d", len(patterns)-2)
	}
	return fmt.Sprintf("%s: %s%s", base, strings.Join(display, ", "), more)
}

// FormatIMToolLine formats one agent tool step (no emoji; aligned with Web getToolTitle).
func FormatIMToolLine(step IMToolStep) string {
	title := imAgentToolTitle(step)
	if title == "" {
		return ""
	}
	if step.Pending {
		return title
	}
	if summary := imToolResultSummary(step); summary != "" {
		return title + " · " + summary
	}
	return title
}

// FormatIMRagPipelineLine formats quick-QA RAG pipeline steps (Web RagPipelineProgress).
func FormatIMRagPipelineLine(step IMToolStep) string {
	toolName := step.ToolName
	query := imGetQueryText(step.Arguments)
	if query == "" {
		query = imGetQueryText(step.Data)
	}

	switch toolName {
	case "query_understand":
		if step.Pending {
			return "Understanding question..."
		}
		return "Question understood"
	case "knowledge_search", "search_knowledge":
		source := imRetrievalSearchSource(step)
		if step.Pending {
			switch source {
			case imRetrievalSourceWeb:
				if query != "" {
					return fmt.Sprintf("Searching web: %s", query)
				}
				return "Searching web..."
			case imRetrievalSourceMixed:
				if query != "" {
					return fmt.Sprintf("Searching knowledge and web: %s", query)
				}
				return "Searching knowledge and web..."
			default:
				if query != "" {
					return fmt.Sprintf("Searching knowledge: %s", query)
				}
				return "Searching knowledge..."
			}
		}
		base := imRetrievalDoneTitle(source, step.Success)
		line := imAppendQueryTitle(base, query)
		if summary := imKnowledgeSearchSummary(step.Data); summary != "" {
			return line + " · " + summary
		}
		return line
	default:
		return ""
	}
}

func imSandboxMutationTitle(step IMToolStep, pending bool) string {
	name := imLocalizedToolName(step.ToolName)
	path := imSandboxFilePath(step)
	stat := imSandboxDiffStat(step)
	base := name
	if pending {
		base = name + "..."
	}
	if path != "" {
		base = imAppendQueryTitle(name, path)
		if pending {
			base += "..."
		}
	}
	if stat != "" {
		return base + " " + stat
	}
	return base
}

func imSandboxFilePath(step IMToolStep) string {
	if step.Data != nil {
		if p, ok := step.Data["path"].(string); ok && strings.TrimSpace(p) != "" {
			return strings.TrimSpace(p)
		}
	}
	if step.Arguments != nil {
		if p, ok := step.Arguments["path"].(string); ok && strings.TrimSpace(p) != "" {
			return strings.TrimSpace(p)
		}
	}
	return ""
}

func imSandboxDiffStat(step IMToolStep) string {
	added := imIntField(step.Arguments, "added_lines")
	removed := imIntField(step.Arguments, "removed_lines")
	if added == 0 && removed == 0 {
		added = imIntField(step.Data, "added_lines")
		removed = imIntField(step.Data, "removed_lines")
	}
	switch {
	case added > 0 && removed > 0:
		return fmt.Sprintf("+%d -%d", added, removed)
	case added > 0:
		return fmt.Sprintf("+%d", added)
	case removed > 0:
		return fmt.Sprintf("-%d", removed)
	default:
		return ""
	}
}

func imAgentToolTitle(step IMToolStep) string {
	if step.Pending {
		switch step.ToolName {
		case "image_analysis":
			return "Analyzing image..."
		case "wiki_search", "wiki_read_page":
			return imLocalizedToolName(step.ToolName) + "..."
		case "write_sandbox_file", "edit_sandbox_file":
			return imSandboxMutationTitle(step, true)
		default:
			return fmt.Sprintf("Calling %s...", imLocalizedToolName(step.ToolName))
		}
	}

	toolName := step.ToolName
	isSearchTool := toolName == "search_knowledge" || toolName == "knowledge_search" || toolName == "wiki_search"
	if isSearchTool {
		base := imToolStatusDescription(step)
		query := imGetQueryText(step.Arguments)
		if query == "" {
			query = imGetQueryText(step.Data)
		}
		return imAppendQueryTitle(base, query)
	}

	if toolName == "web_search" {
		base := imToolStatusDescription(step)
		return imAppendQueryTitle(base, imGetWebSearchQuery(step))
	}

	if toolName == "grep_chunks" {
		base := imToolStatusDescription(step)
		return imAppendPatternsTitle(base, imGetGrepPatternsFromStep(step))
	}

	if toolName == "wiki_read_page" {
		pageLabel := ""
		if step.Data != nil {
			if title, ok := step.Data["title"].(string); ok {
				pageLabel = strings.TrimSpace(title)
			}
		}
		if pageLabel == "" {
			pageLabel = imGetWikiPageText(step.Arguments)
		}
		if pageLabel == "" {
			pageLabel = imGetWikiPageText(step.Data)
		}
		base := imToolStatusDescription(step)
		return imAppendQueryTitle(base, pageLabel)
	}

	if toolName == "write_sandbox_file" || toolName == "edit_sandbox_file" {
		return imSandboxMutationTitle(step, false)
	}

	if summary := imToolHeaderSummary(step); summary != "" {
		return summary
	}
	return imToolStatusDescription(step)
}

func imToolStatusDescription(step IMToolStep) string {
	success := step.Success
	toolName := step.ToolName

	switch toolName {
	case "search_knowledge", "knowledge_search":
		if success {
			return "Search knowledge"
		}
		return "Knowledge search failed"
	case "wiki_search", "wiki_read_page":
		name := imLocalizedToolName(toolName)
		if success {
			return name
		}
		return fmt.Sprintf("Calling %s failed", name)
	case "web_search":
		if success {
			return "Search web"
		}
		return "Web search failed"
	case "grep_chunks":
		if success {
			return "Search keywords"
		}
		return "Keyword search failed"
	case "get_document_info":
		if success {
			return "Get document details"
		}
		return "Getting document details failed"
	case "get_document_content", "wiki_read_source_doc", "read_document":
		if success {
			return "Read document"
		}
		return "Reading document failed"
	case "list_documents":
		if success {
			return "Browse documents"
		}
		return "Browsing documents failed"
	case "thinking":
		if success {
			return "Thinking complete"
		}
		return "Thinking failed"
	case "todo_write":
		if success {
			return "Update task list"
		}
		return "Updating task list failed"
	case "image_analysis":
		if success {
			return "Image analyzed"
		}
		return "Image analysis failed"
	case "query_understand":
		if success {
			return "Question understood"
		}
		return fmt.Sprintf("Calling %s failed", imLocalizedToolName(toolName))
	default:
		name := imLocalizedToolName(toolName)
		if success {
			return fmt.Sprintf("Call %s", name)
		}
		return fmt.Sprintf("Calling %s failed", name)
	}
}

func imToolHeaderSummary(step IMToolStep) string {
	if step.Pending || !step.Success {
		return ""
	}
	toolName := step.ToolName
	data := step.Data

	switch toolName {
	case "search_knowledge", "knowledge_search":
		return ""
	case "get_document_info":
		if data != nil {
			if title, ok := data["title"].(string); ok && strings.TrimSpace(title) != "" {
				return fmt.Sprintf("Get document: %s", strings.TrimSpace(title))
			}
		}
	case "list_knowledge_chunks", "read_document":
		if data != nil {
			if question, ok := data["faq_question"].(string); ok && strings.TrimSpace(question) != "" {
				return fmt.Sprintf("View FAQ: %s", strings.TrimSpace(question))
			}
			if _, ok := data["fetched_chunks"]; ok {
				title := "document"
				if t, ok := data["knowledge_title"].(string); ok && strings.TrimSpace(t) != "" {
					title = strings.TrimSpace(t)
				} else if id, ok := data["knowledge_id"].(string); ok && strings.TrimSpace(id) != "" {
					title = strings.TrimSpace(id)
				}
				return fmt.Sprintf("View %s", title)
			}
		}
	}
	return ""
}

func imToolResultSummary(step IMToolStep) string {
	if step.Pending || !step.Success {
		return ""
	}
	switch step.ToolName {
	case "search_knowledge", "knowledge_search":
		return imKnowledgeSearchSummary(step.Data)
	case "web_search":
		return imWebSearchSummary(step.Data)
	case "grep_chunks":
		return imGrepSearchSummary(step.Data)
	case "list_knowledge_chunks", "read_document":
		return imKnowledgeChunksSummary(step.Data)
	case "list_documents":
		return imDocumentListSummary(step.Data)
	default:
		return briefToolSummary(step.Output)
	}
}

const (
	imRetrievalSourceKnowledge = "knowledge"
	imRetrievalSourceWeb       = "web"
	imRetrievalSourceMixed     = "mixed"
)

func imRetrievalSearchSource(step IMToolStep) string {
	if source := imSearchSourceFromData(step.Data); source != "" {
		return source
	}
	if step.Arguments != nil {
		if source, ok := step.Arguments["search_source"].(string); ok && source != "" {
			return source
		}
	}
	return imRetrievalSourceKnowledge
}

func imSearchSourceFromData(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	if source, ok := data["search_source"].(string); ok {
		return strings.TrimSpace(source)
	}
	return ""
}

func imRetrievalDoneTitle(source string, success bool) string {
	switch source {
	case imRetrievalSourceWeb:
		if success {
			return "Web search"
		}
		return "Web search failed"
	case imRetrievalSourceMixed:
		if success {
			return "Search knowledge and web"
		}
		return "Search failed"
	default:
		if success {
			return "Search knowledge"
		}
		return "Knowledge search failed"
	}
}

func imIntField(data map[string]interface{}, key string) int {
	if data == nil {
		return 0
	}
	switch v := data[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func imKnowledgeSearchSummary(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	count := imResultCount(data)
	if count == 0 {
		return "No matching content found"
	}
	source := imSearchSourceFromData(data)
	webCount := imIntField(data, "web_count")
	docCount := imIntField(data, "doc_count")
	if source == imRetrievalSourceWeb || (webCount > 0 && docCount == 0) {
		return fmt.Sprintf("Found %d web pages", count)
	}
	if kbCounts, ok := data["kb_counts"].(map[string]interface{}); ok && len(kbCounts) > 0 {
		return fmt.Sprintf("Found %d results from %d files", count, len(kbCounts))
	}
	if source == imRetrievalSourceMixed && docCount > 0 && webCount > 0 {
		return fmt.Sprintf("Found %d results (%d documents, %d web pages)", count, docCount, webCount)
	}
	return fmt.Sprintf("Found %d results", count)
}

func imWebSearchSummary(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	count := imResultCount(data)
	if count == 0 {
		return ""
	}
	return fmt.Sprintf("Found %d web search results", count)
}

func imGrepSearchSummary(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	totalChunks := 0
	if v, ok := data["total_matches"].(float64); ok {
		totalChunks = int(v)
	} else if v, ok := data["total_matches"].(int); ok {
		totalChunks = v
	}
	if totalChunks == 0 {
		return "No matching content found"
	}
	docCount := imGrepDocumentCount(data)
	return fmt.Sprintf("Found %d matching chunks from %d documents", totalChunks, docCount)
}

func imGrepDocumentCount(data map[string]interface{}) int {
	if v, ok := data["document_count"].(float64); ok && v >= 0 {
		return int(v)
	}
	if v, ok := data["document_count"].(int); ok && v >= 0 {
		return v
	}
	if results, ok := data["knowledge_results"].([]interface{}); ok && len(results) > 0 {
		return len(results)
	}
	if results, ok := data["chunk_results"].([]interface{}); ok && len(results) > 0 {
		return len(results)
	}
	return 0
}

func imDocumentListSummary(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	total := imNumericValue(data["total_docs"])
	listed := 0
	if docs, ok := data["documents"].([]interface{}); ok {
		listed = len(docs)
	}
	if total == 0 && listed == 0 {
		return "No documents in the knowledge base"
	}
	return fmt.Sprintf("Listed %d of %d documents", listed, total)
}

func imKnowledgeChunksSummary(data map[string]interface{}) string {
	if data == nil {
		return ""
	}
	fetched, ok := data["fetched_chunks"]
	if !ok {
		return ""
	}
	fetchedN := imNumericValue(fetched)
	totalN := imNumericValue(data["total_chunks"])
	summary := fmt.Sprintf("Loaded %d of %v chunks", fetchedN, formatIMOptionalInt(totalN, data["total_chunks"]))
	pageSize := imNumericValue(data["page_size"])
	if totalN > pageSize && pageSize > 0 {
		page := imNumericValue(data["page"])
		if page <= 0 {
			page = 1
		}
		summary += fmt.Sprintf(" · page %d, %d per page", page, pageSize)
	}
	return summary
}

func imResultCount(data map[string]interface{}) int {
	if results, ok := data["results"].([]interface{}); ok && len(results) > 0 {
		return len(results)
	}
	if count, ok := data["count"].(float64); ok && count > 0 {
		return int(count)
	}
	if count, ok := data["count"].(int); ok && count > 0 {
		return count
	}
	return 0
}

func imNumericValue(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func formatIMOptionalInt(n int, raw any) string {
	if raw == nil {
		return "?"
	}
	if _, ok := raw.(string); ok && n == 0 {
		return "?"
	}
	if n == 0 {
		return "?"
	}
	return fmt.Sprintf("%d", n)
}

func renderIMToolSteps(steps []IMToolStep, format func(IMToolStep) string) string {
	if len(steps) == 0 {
		return ""
	}
	var b strings.Builder
	for _, step := range steps {
		line := format(step)
		if line == "" {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func mergeIMNarrativeAndTools(narrative string, toolLines string) string {
	narrative = strings.TrimSpace(narrative)
	toolLines = strings.TrimSpace(toolLines)
	switch {
	case narrative != "" && toolLines != "":
		return narrative + "\n" + toolLines
	case toolLines != "":
		return toolLines
	default:
		return narrative
	}
}

func upsertIMToolStep(steps *[]IMToolStep, index map[string]int, id string, update func(*IMToolStep)) {
	if i, ok := index[id]; ok {
		update(&(*steps)[i])
		return
	}
	step := IMToolStep{ToolCallID: id}
	update(&step)
	index[id] = len(*steps)
	*steps = append(*steps, step)
}
