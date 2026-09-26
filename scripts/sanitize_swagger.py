"""Remove retired routes and non-English metadata from generated Swagger docs."""

import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
CJK = re.compile(r"[\u3400-\u9fff]")
OLD_MODULE = "github_com_Tencent_WeKnora"
NEW_MODULE = "github_com_ai_tool_collection_WeKnora"


def clean(value, key=""):
    if isinstance(value, dict):
        return {clean(k): clean(v, k) for k, v in value.items()
                if k not in {"weknoracloud"}}
    if isinstance(value, list):
        return [clean(item, key) for item in value
                if not (key == "enum" and item == "tencent_vectordb")]
    if isinstance(value, str):
        value = value.replace(OLD_MODULE, NEW_MODULE)
        value = value.replace("Tencent VectorDB", "unsupported legacy vector database")
        value = value.replace("TencentVectorDBRetrieverEngineType", "UnsupportedLegacyRetrieverEngineType")
        value = value.replace("WeKnoraCloud", "retired cloud service")
        if CJK.search(value):
            if key in {"description", "summary", "title", "name", "example"}:
                return ""
            return re.sub(r"[\u3400-\u9fff]+", "", value).strip()
        return value
    return value


doc = clean(json.loads((DOCS / "swagger.json").read_text(encoding="utf-8")))
doc["info"].update({
    "description": "Knowledge Hub API reference",
    "title": "Knowledge Hub API",
    "contact": {"name": "Knowledge Hub", "url": "https://github.com/ai-tool-collection/WeKnora"},
})
for path in list(doc["paths"]):
    if "weknoracloud" in path.lower():
        del doc["paths"][path]
for name in list(doc.get("definitions", {})):
    if "retired cloud service" in name:
        del doc["definitions"][name]

for path, operations in doc["paths"].items():
    for method, operation in operations.items():
        if isinstance(operation, dict) and not operation.get("summary"):
            operation["summary"] = f"{method.upper()} {path}"

json_text = json.dumps(doc, ensure_ascii=False, indent=4) + "\n"
(DOCS / "swagger.json").write_text(json_text, encoding="utf-8", newline="\n")
# JSON is valid YAML 1.2, so one dependency-free serialization keeps the
# registered Go document, JSON file, and YAML file in lockstep.
(DOCS / "swagger.yaml").write_text(json_text, encoding="utf-8", newline="\n")

go = '''// Package docs contains the Knowledge Hub Swagger specification.
package docs

import (
    _ "embed"
    "github.com/swaggo/swag"
)

//go:embed swagger.json
var docTemplate string

var SwaggerInfo = &swag.Spec{
    Version:          "1.0",
    Host:             "",
    BasePath:         "/api/v1",
    Schemes:          []string{},
    Title:            "Knowledge Hub API",
    Description:      "Knowledge Hub API reference",
    InfoInstanceName: "swagger",
    SwaggerTemplate:  docTemplate,
    LeftDelim:        "{{",
    RightDelim:       "}}",
}

func init() {
    swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
'''
(DOCS / "docs.go").write_text(go, encoding="utf-8", newline="\n")
