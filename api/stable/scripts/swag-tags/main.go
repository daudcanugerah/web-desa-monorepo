// Package main is a swag post-processor that injects a top-level `tags:`
// array into the generated swagger.{json,yaml} files. swag v1.x does NOT
// generate this from @Tags annotations — only per-operation tags — so the
// Swagger UI sidebar ends up empty without it.
//
// Usage:
//
//	go run ./scripts/swag-tags/main.go
//
// Reads docs/swagger.json, walks all `tags: [...]` arrays under `paths.*`,
// collects the unique set, looks up descriptions from a static map, and
// writes a top-level `tags: [...]` array back into the file. Same for YAML.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// tagDescriptions maps tag names used in handler @Tags annotations to human
// descriptions shown in the Swagger UI sidebar. Order here controls the
// sidebar order in most Swagger UI renderers.
var tagDescriptions = []struct {
	Name        string
	Description string
}{
	{"health", "Service health check"},
	{"auth", "Authentication (login, refresh, password reset)"},
	{"users", "User account management"},
	{"roles", "RBAC role and permission management"},
	{"desa", "Village profile"},
	{"berita", "News / articles"},
	{"berita-categories", "News category vocabulary"},
	{"berita-upload", "Quill editor media upload"},
	{"banner", "Banner carousel"},
	{"umkm", "Small business directory"},
	{"umkm-categories", "UMKM category vocabulary"},
	{"fasilitas", "Village facilities (with geolocation)"},
	{"fasilitas-categories", "Fasilitas category vocabulary"},
	{"ppid", "Public Information Disclosure (PPID) documents"},
	{"ppid-requests", "PPID document request workflow"},
	{"ppid-categories", "PPID category vocabulary"},
	{"struktur", "Organization chart members"},
	{"profile", "CMS-style profile sections"},
	{"infographic", "Embedded Metabase dashboards/questions"},
	{"infographic-categories", "Infographic category vocabulary"},
	{"files", "Public file serving"},
}

var tagDescriptionsByName = func() map[string]string {
	m := map[string]string{}
	for _, t := range tagDescriptions {
		m[t.Name] = t.Description
	}
	return m
}()

func main() {
	for _, name := range []string{"docs/swagger.json", "docs/swagger.yaml"} {
		if err := process(name); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("✓ %s\n", name)
	}
}

func process(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	ext := filepath.Ext(path)
	var root map[string]any
	if ext == ".json" {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse json: %w", err)
		}
	} else {
		if err := yaml.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("parse yaml: %w", err)
		}
	}

	tags := collectTags(root)

	// Build ordered tag list (by description map, then alphabetical fallback)
	seen := map[string]bool{}
	ordered := []map[string]any{}
	for _, t := range tagDescriptions {
		if !tags[t.Name] {
			continue
		}
		ordered = append(ordered, map[string]any{
			"name":        t.Name,
			"description": t.Description,
		})
		seen[t.Name] = true
	}
	leftover := []string{}
	for name := range tags {
		if !seen[name] {
			leftover = append(leftover, name)
		}
	}
	sort.Strings(leftover)
	for _, name := range leftover {
		ordered = append(ordered, map[string]any{
			"name":        name,
			"description": name,
		})
	}

	root["tags"] = ordered

	var out []byte
	if ext == ".json" {
		out, err = json.MarshalIndent(root, "", "    ")
	} else {
		out, err = yaml.Marshal(root)
	}
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(path, out, 0644)
}

func collectTags(root map[string]any) map[string]bool {
	out := map[string]bool{}
	paths, _ := root["paths"].(map[string]any)
	for _, v := range paths {
		ops, _ := v.(map[string]any)
		for _, op := range ops {
			opMap, _ := op.(map[string]any)
			tagList, _ := opMap["tags"].([]any)
			for _, t := range tagList {
				if s, ok := t.(string); ok && s != "" {
					out[s] = true
				}
			}
		}
	}
	return out
}