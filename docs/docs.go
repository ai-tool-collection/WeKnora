// Package docs contains the Knowledge Hub Swagger specification.
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
