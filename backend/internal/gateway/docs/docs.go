package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "NeuralOps API Gateway",
        "version": "1.0",
        "description": "Single entry point for NeuralOps observability platform"
    },
    "basePath": "/api/v1",
    "paths": {
        "/dashboard/overview": {
            "get": {
                "summary": "Dashboard overview",
                "responses": { "200": { "description": "Aggregated dashboard metrics" } }
            }
        },
        "/chat/query": {
            "post": {
                "summary": "AI chat query (SSE stream)",
                "responses": { "200": { "description": "Server-sent events stream" } }
            }
        },
        "/search/logs": {
            "post": { "summary": "Structured log search", "responses": { "200": { "description": "Search results" } } }
        },
        "/search/semantic": {
            "post": { "summary": "Semantic/hybrid search", "responses": { "200": { "description": "Search results" } } }
        },
        "/search/transactions": {
            "post": { "summary": "Transaction journey search", "responses": { "200": { "description": "Transaction list" } } }
        },
        "/incidents": {
            "get": { "summary": "List incidents", "responses": { "200": { "description": "Incident list" } } }
        }
    }
}`

type swaggerInfo struct{}

func (s *swaggerInfo) ReadDoc() string { return docTemplate }

func init() {
	swag.Register(swag.Name, &swaggerInfo{})
}
