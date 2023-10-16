// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/bridge": {
            "get": {
                "description": "Get the current bridge and its connected interface",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bridge"
                ],
                "summary": "Get current bridge and connected interface",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/main.BridgeResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/bridge/{bridge_name}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bridge"
                ],
                "summary": "Retrieve bridge status",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Bridge existed",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Bridge not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "System error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "post": {
                "description": "Add a new bridge with the given name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "bridge"
                ],
                "summary": "Add a new bridge",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Bridge created successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid bridge name",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/interface": {
            "post": {
                "description": "Add a new interface between two bridges",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "interface"
                ],
                "summary": "Add a new interface",
                "parameters": [
                    {
                        "description": "Interface request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.InterfaceRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Interface added successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid request body",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/slice/{bridge_name}": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "slice"
                ],
                "summary": "Add slice on interface",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Slice request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.SliceRequest"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Slice Installed",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/slice/{bridge_name}/{slice_sd}}": {
            "delete": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "slice"
                ],
                "summary": "Del slice on interface",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Slice SD identifier",
                        "name": "slice_sd",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Slice deletion successful",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/vxlan/{bridge_name}": {
            "post": {
                "description": "Add a new bridge with the given name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vxlan"
                ],
                "summary": "Add a new bridge",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Vxlan Interface request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/main.VxlanInterfaceRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Bridge created successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid bridge name",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            },
            "delete": {
                "description": "Delete a exist bridge with the given name",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vxlan"
                ],
                "summary": "Delete bridge",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Bridge delete successfully",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/v1/vxlan/{bridge_name}/activate": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "vxlan"
                ],
                "summary": "[Deprecated] Activate vxlan bridge",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Bridge name",
                        "name": "bridge_name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "Bridge Activated",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Invalid bridge name",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "main.BridgeResponse": {
            "type": "object",
            "properties": {
                "bridge": {
                    "type": "string"
                },
                "interface": {
                    "type": "string"
                }
            }
        },
        "main.InterfaceRequest": {
            "type": "object",
            "properties": {
                "bridge1": {
                    "type": "string"
                },
                "bridge2": {
                    "type": "string"
                }
            }
        },
        "main.SliceRequest": {
            "type": "object",
            "properties": {
                "DstIP": {
                    "type": "string"
                },
                "FlowRate": {
                    "type": "integer"
                },
                "SliceSD": {
                    "type": "string"
                },
                "SrcIP": {
                    "type": "string"
                }
            }
        },
        "main.VxlanInterfaceRequest": {
            "type": "object",
            "properties": {
                "bindInterface": {
                    "type": "string"
                },
                "localBrIp": {
                    "type": "string"
                },
                "remoteIp": {
                    "description": "LocalBridgeName\t\tstring\t` + "`" + `json:\"localBrName\"` + "`" + `",
                    "type": "string"
                },
                "vxlanId": {
                    "type": "string"
                },
                "vxlanInterface": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Bridge API",
	Description:      "API endpoints for managing bridges and interfaces.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
