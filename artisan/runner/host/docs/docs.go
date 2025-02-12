// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "gatblau",
            "url": "http://onix.gatblau.org/",
            "email": "onix@gatblau.org"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/flow": {
            "post": {
                "description": "Execute a flow from the definition passed in the payload.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Runner"
                ],
                "summary": "Execute an Artisan flow",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/host/{cmd-key}": {
            "post": {
                "description": "Trigger a new build to create artisan package from the vulnerabilty scanned csv report passed in the payload.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Runner"
                ],
                "summary": "Build patching artisan package",
                "parameters": [
                    {
                        "type": "string",
                        "description": "the unique key of the command to retrieve",
                        "name": "cmd-key",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "422": {
                        "description": "Unprocessable Entity",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/webhook/{flow-key}/push": {
            "post": {
                "description": "Connect to CMDB and retrieves a flow using configuration item natural key passed in flow-key from CMDB",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Runner"
                ],
                "summary": "Retrieve a configured flow from CMDB and execute it.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "the unique key of the flow specification in cmdb",
                        "name": "flow-key",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.0.4",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Artisan Host Runner",
	Description:      "Run Artisan packages with in a host",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
