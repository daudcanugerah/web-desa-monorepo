// Package main is the entry point for the Desa API binary.
//
// The //-comments above main() are the top-level swag annotations read by
// `swag init` to populate the OpenAPI document's `info` block.
//
// Regenerate the OpenAPI spec with:
//
//	swag init -g main.go -o docs --parseDependency --parseInternal
package main

import (
	"webdesa/api/cmd"
	"os"
)

// Desa API
//
// @title                       Desa API
// @version                     1.0
// @description                 REST API for the Desa (village) information system. Implements Clean Architecture with strict dependency rules, JWT (HS256) authentication, Casbin RBAC, PostgreSQL persistence, and rate limiting.
// @host                        localhost:8080
// @BasePath                    /api/v1
// @schemes                     http https
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT bearer token. Format: `Bearer {token}`
func main() {
	i := cmd.Execute()
	os.Exit(i)
}