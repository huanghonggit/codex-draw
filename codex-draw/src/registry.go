package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type modelProfile struct {
	SchemaVersion       int      `json:"schema_version"`
	ID                  string   `json:"id"`
	Adapter             string   `json:"adapter"`
	Operations          []string `json:"operations"`
	Formats             []string `json:"formats"`
	ExperimentalFormats []string `json:"experimental_formats"`
	Priority            int      `json:"priority"`
	Notes               string   `json:"notes"`
}
type scenarioProfile struct {
	SchemaVersion     int      `json:"schema_version"`
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Operation         string   `json:"operation"`
	DefaultFormat     string   `json:"default_format"`
	ReferenceRequired bool     `json:"reference_required"`
	ArticleCases      []int    `json:"article_cases"`
	Inputs            []string `json:"inputs"`
	Prompt            string   `json:"-"`
	Checklist         string   `json:"-"`
}
type registry struct {
	Models    []modelProfile    `json:"models"`
	Scenarios []scenarioProfile `json:"scenarios"`
}

var idPattern = regexp.MustCompile("^[a-z0-9][a-z0-9._-]{0,99}$")

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
func registryPath() string {
	if p := os.Getenv("CODEX_DRAW_REGISTRY"); p != "" {
		return p
	}
	exe, _ := os.Executable()
	for _, p := range []string{filepath.Join(filepath.Dir(exe), "..", "registry"), "registry", filepath.Join("codex-draw", "registry"), filepath.Join("..", "registry")} {
		if st, e := os.Stat(p); e == nil && st.IsDir() {
			return p
		}
	}
	return "registry"
}
func strictJSON(p string, dst any) error {
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if e != nil || len(data) > 1<<20 {
		return fmt.Errorf("registry file unreadable or too large: %s", p)
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return fmt.Errorf("invalid JSON object: %s", p)
	}
	required := []string{}
	switch dst.(type) {
	case *modelProfile:
		required = []string{"schema_version", "id", "adapter", "operations", "formats", "experimental_formats", "priority", "notes"}
	case *scenarioProfile:
		required = []string{"schema_version", "id", "title", "description", "operation", "default_format", "reference_required", "article_cases", "inputs"}
	}
	for _, name := range required {
		v, ok := fields[name]
		if !ok || string(v) == "null" {
			return fmt.Errorf("missing or null %s in %s", name, p)
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e = d.Decode(dst); e != nil {
		return fmt.Errorf("%s: %w", p, e)
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("%s: trailing JSON", p)
	}
	return nil
}
func loadRegistry(root string) (*registry, error) {
	r := &registry{}
	seen := map[string]bool{}
	files, e := filepath.Glob(filepath.Join(root, "models", "*.json"))
	if e != nil {
		return nil, e
	}
	for _, p := range files {
		var m modelProfile
		if e := strictJSON(p, &m); e != nil {
			return nil, e
		}
		if m.SchemaVersion != 1 || m.Priority < 0 || filepath.Base(p) != m.ID+".json" || !idPattern.MatchString(m.ID) || seen[m.ID] || m.Adapter != "openai-images-v1" || len(m.Operations) == 0 || len(m.Formats) == 0 || m.Notes == "" {
			return nil, fmt.Errorf("invalid model profile: %s", p)
		}
		for _, o := range m.Operations {
			if o != "generate" && o != "edit" {
				return nil, fmt.Errorf("unknown operation in %s", p)
			}
		}
		for _, f := range m.Formats {
			if !contains([]string{"png", "jpeg", "gif"}, f) {
				return nil, fmt.Errorf("adapter cannot encode %s in %s", f, p)
			}
		}
		for _, f := range m.ExperimentalFormats {
			if !contains(m.Formats, f) {
				return nil, fmt.Errorf("experimental format not supported in %s", p)
			}
		}
		seen[m.ID] = true
		r.Models = append(r.Models, m)
	}
	dirs, e := os.ReadDir(filepath.Join(root, "scenarios"))
	if e != nil {
		return nil, e
	}
	seen = map[string]bool{}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		dir := filepath.Join(root, "scenarios", d.Name())
		var s scenarioProfile
		if e := strictJSON(filepath.Join(dir, "scenario.json"), &s); e != nil {
			return nil, e
		}
		if s.SchemaVersion != 1 || !idPattern.MatchString(s.ID) || s.ID != d.Name() || seen[s.ID] || s.Title == "" || s.Description == "" || len(s.Inputs) == 0 || !contains([]string{"generate", "edit", "handoff"}, s.Operation) || !contains([]string{"png", "jpeg", "gif", "svg"}, s.DefaultFormat) {
			return nil, fmt.Errorf("invalid scenario: %s", dir)
		}
		if s.DefaultFormat == "svg" && s.Operation != "handoff" {
			return nil, fmt.Errorf("SVG needs handoff: %s", dir)
		}
		if s.ReferenceRequired && s.Operation != "edit" {
			return nil, fmt.Errorf("reference-required scenario must use edit: %s", dir)
		}
		for _, input := range s.Inputs {
			if strings.TrimSpace(input) == "" {
				return nil, fmt.Errorf("empty scenario input: %s", dir)
			}
		}
		for _, n := range s.ArticleCases {
			if n < 1 || n > 10 {
				return nil, fmt.Errorf("invalid article mapping: %s", dir)
			}
		}
		b, e := os.ReadFile(filepath.Join(dir, "prompt.md"))
		if e != nil {
			return nil, e
		}
		s.Prompt = string(b)
		b, e = os.ReadFile(filepath.Join(dir, "checklist.md"))
		if e != nil {
			return nil, e
		}
		s.Checklist = string(b)
		if !strings.Contains(s.Prompt, "{{brief}}") || strings.TrimSpace(s.Checklist) == "" {
			return nil, fmt.Errorf("missing brief placeholder or checklist: %s", dir)
		}
		seen[s.ID] = true
		r.Scenarios = append(r.Scenarios, s)
	}
	if len(r.Models) == 0 || len(r.Scenarios) == 0 {
		return nil, errors.New("registry needs models and scenarios")
	}
	sort.Slice(r.Models, func(i, j int) bool {
		if r.Models[i].Priority != r.Models[j].Priority {
			return r.Models[i].Priority < r.Models[j].Priority
		}
		return r.Models[i].ID < r.Models[j].ID
	})
	return r, nil
}
func (r *registry) model(id string) (modelProfile, bool) {
	for _, m := range r.Models {
		if m.ID == id {
			return m, true
		}
	}
	return modelProfile{}, false
}
func checkModelCapability(model, operation, format string, allow bool) error {
	r, e := loadRegistry(registryPath())
	if e != nil {
		return e
	}
	m, ok := r.model(model)
	if !ok {
		return fmt.Errorf("model %q is unregistered; add registry/models/<id>.json", model)
	}
	if format == "" {
		format = "png"
	}
	if !contains(m.Operations, operation) || !contains(m.Formats, format) {
		return fmt.Errorf("model %s does not declare %s/%s", model, operation, format)
	}
	if contains(m.ExperimentalFormats, format) && !allow {
		return fmt.Errorf("%s/%s is provider-dependent; use --allow-experimental only after provider support is established", model, format)
	}
	return nil
}
