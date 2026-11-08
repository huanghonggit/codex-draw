package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func syntheticPNG() []byte {
	var b bytes.Buffer
	im := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	im.Set(1, 1, color.NRGBA{R: 230, A: 255})
	_ = png.Encode(&b, im)
	return b.Bytes()
}
func syntheticGIF() []byte {
	var b bytes.Buffer
	p := color.Palette{color.Black, color.White}
	a := image.NewPaletted(image.Rect(0, 0, 2, 2), p)
	z := image.NewPaletted(image.Rect(0, 0, 2, 2), p)
	_ = gif.EncodeAll(&b, &gif.GIF{Image: []*image.Paletted{a, z}, Delay: []int{10, 20}, LoopCount: 0})
	return b.Bytes()
}
func isolatedSettings(t *testing.T) {
	t.Helper()
	for _, n := range []string{"APPDATA", "HOME", "USERPROFILE", "XDG_CONFIG_HOME"} {
		t.Setenv(n, t.TempDir())
	}
	t.Setenv("CODEX_API_KEY", "")
	t.Setenv("CODEX_API_URL", "")
}
func copyRegistry(t *testing.T) string {
	t.Helper()
	source := registryPath()
	dest := t.TempDir()
	e := filepath.WalkDir(source, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		rel, _ := filepath.Rel(source, p)
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		return os.WriteFile(target, b, 0644)
	})
	if e != nil {
		t.Fatal(e)
	}
	return dest
}
func TestInspectRealOutput(t *testing.T) {
	m, e := inspectBytes(syntheticGIF())
	if e != nil || m["frames"] != 2 || m["duration_ms"] != 300 || m["width"] != 2 || m["loop_count"] != 0 || m["first_last_equal"] != true {
		t.Fatal(m, e)
	}
	m, e = inspectBytes(syntheticPNG())
	if e != nil || m["has_transparency"] != true || m["height"] != 2 {
		t.Fatal(m, e)
	}
	for _, b := range [][]byte{[]byte("GIF89a"), []byte("<html>error</html>"), syntheticPNG()[:30]} {
		if _, e = inspectBytes(b); e == nil {
			t.Fatal("malformed output accepted")
		}
	}
}
func TestSaveValidatesAllBeforeWritingAndRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	if _, e := saveImages([][]byte{syntheticPNG(), []byte("bad")}, []string{a, b}, false); e == nil {
		t.Fatal("bad image accepted")
	}
	if _, e := os.Stat(a); !os.IsNotExist(e) {
		t.Fatal("partially wrote malformed response")
	}
	if _, e := saveImages([][]byte{syntheticPNG()}, []string{a}, false); e != nil {
		t.Fatal(e)
	}
	if _, e := saveImages([][]byte{syntheticPNG()}, []string{a}, false); e == nil {
		t.Fatal("overwrote existing output")
	}
	if e := validateImageOutputs([][]byte{syntheticPNG()}, []string{"a.gif"}); e == nil {
		t.Fatal("format mismatch accepted")
	}
}
func TestRegistryExtensionsAndRejectUnknownFields(t *testing.T) {
	dir := copyRegistry(t)
	m := modelProfile{SchemaVersion: 1, ID: "new-provider-image", Adapter: "openai-images-v1", Operations: []string{"generate"}, Formats: []string{"png"}, ExperimentalFormats: []string{}, Priority: 1, Notes: "Test only"}
	b, _ := json.Marshal(m)
	p := filepath.Join(dir, "models", m.ID+".json")
	_ = os.WriteFile(p, b, 0644)
	r, e := loadRegistry(dir)
	if e != nil || r.Models[0].ID != m.ID {
		t.Fatal("data-only model extension failed", e)
	}
	t.Setenv("CODEX_DRAW_REGISTRY", dir)
	if e := checkModelCapability(m.ID, "generate", "png", false); e != nil {
		t.Fatal(e)
	}
	if e := checkModelCapability(m.ID, "edit", "png", false); e == nil {
		t.Fatal("unsupported edit allowed")
	}
	b = bytes.Replace(b, []byte("\"priority\":"), []byte("\"typo\":"), 1)
	_ = os.WriteFile(p, b, 0644)
	if _, e := loadRegistry(dir); e == nil {
		t.Fatal("unknown field accepted")
	}
}
func TestScenarioExtensionAndMissingTemplate(t *testing.T) {
	dir := copyRegistry(t)
	p := filepath.Join(dir, "scenarios", "new-poster")
	_ = os.MkdirAll(p, 0755)
	s := scenarioProfile{SchemaVersion: 1, ID: "new-poster", Title: "Poster", Description: "A new poster", Operation: "generate", DefaultFormat: "png", ArticleCases: []int{}, Inputs: []string{"brief"}}
	b, _ := json.Marshal(s)
	_ = os.WriteFile(filepath.Join(p, "scenario.json"), b, 0644)
	_ = os.WriteFile(filepath.Join(p, "prompt.md"), []byte("Make poster: {{brief}}"), 0644)
	_ = os.WriteFile(filepath.Join(p, "checklist.md"), []byte("Check title"), 0644)
	r, e := loadRegistry(dir)
	if e != nil || len(r.Scenarios) != 13 {
		t.Fatal(e)
	}
	_ = os.WriteFile(filepath.Join(p, "prompt.md"), []byte("missing placeholder"), 0644)
	if _, e = loadRegistry(dir); e == nil {
		t.Fatal("invalid template accepted")
	}
}
func TestPlanOfflineAndSVGDoesNotGenerate(t *testing.T) {
	isolatedSettings(t)
	calls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(500) }))
	defer s.Close()
	t.Setenv("CODEX_API_URL", s.URL)
	if e := runWorkflow("plan", []string{"--scenario", "ui-mockup", "--brief", "A calendar"}); e != nil {
		t.Fatal(e)
	}
	if e := runWorkflow("plan", []string{"--scenario", "vector-icons", "--brief", "weather"}); e != nil {
		t.Fatal(e)
	}
	if e := runWorkflow("draw", []string{"--scenario", "vector-icons", "--brief", "weather"}); e == nil {
		t.Fatal("unsupported SVG must not report success")
	}
	if calls != 0 {
		t.Fatal("offline plan accessed network")
	}
}
func TestCapabilityGateAndReferenceRequirements(t *testing.T) {
	if e := checkModelCapability("gpt-image-2.6-sunburst", "edit", "gif", false); e == nil {
		t.Fatal("unverified GIF allowed")
	}
	if e := checkModelCapability("gpt-image-2.6-sunburst", "edit", "gif", true); e != nil {
		t.Fatal(e)
	}
	if e := runWorkflow("plan", []string{"--scenario", "character-loop", "--brief", "blink"}); e == nil {
		t.Fatal("missing reference accepted")
	}
	if e := runWorkflow("plan", []string{"--scenario", "ui-mockup", "--brief", "x", "--model", "unknown"}); e == nil {
		t.Fatal("unknown model accepted")
	}
}
func TestDefaultOneAttempt(t *testing.T) {
	isolatedSettings(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(524) }))
	defer server.Close()
	t.Setenv("CODEX_API_URL", server.URL)
	t.Setenv("CODEX_API_KEY", "synthetic-test-key")
	args := commonArgs{model: defaultModel, size: defaultSize, quality: defaultQuality, n: 1, outDir: t.TempDir(), maxAttempts: 1, timeout: time.Second}
	_, e := generate("test", "x.png", args)
	if e == nil || calls != 1 || strings.Contains(e.Error(), "synthetic-test-key") {
		t.Fatal(calls, e)
	}
}
func TestImportPreservesExistingConfiguration(t *testing.T) {
	isolatedSettings(t)
	store, e := defaultConfigStore()
	if e != nil {
		t.Fatal(e)
	}
	_ = os.MkdirAll(filepath.Dir(store.path), 0700)
	_ = os.WriteFile(store.path, []byte("existing"), 0600)
	if e := importImage2(); e == nil {
		t.Fatal("must not overwrite")
	}
	b, _ := os.ReadFile(store.path)
	if string(b) != "existing" {
		t.Fatal("config changed")
	}
}

func TestScenarioDrawEndToEnd(t *testing.T) {
	isolatedSettings(t)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]any
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Error("bad payload")
		}
		if body["model"] != defaultModel || body["output_format"] != "png" || !strings.Contains(body["prompt"].(string), "calendar") {
			t.Error("scenario was not rendered into request", body)
		}
		if _, ok := body["frame_count"]; ok {
			t.Error("GIF fields leaked into PNG request")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"b64_json": base64.StdEncoding.EncodeToString(syntheticPNG())}}})
	}))
	defer server.Close()
	t.Setenv("CODEX_API_URL", server.URL)
	t.Setenv("CODEX_API_KEY", "synthetic-scenario-test-key")
	out := filepath.Join(t.TempDir(), "ui.png")
	if e := runWorkflow("draw", []string{"--scenario", "ui-mockup", "--brief", "calendar", "--model", defaultModel, "--out", out}); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(out)
	if e != nil || !bytes.Equal(b, syntheticPNG()) || calls != 1 {
		t.Fatal("missing output or duplicate call", calls, e)
	}
}
func TestGIFDisposalRestoresCanvas(t *testing.T) {
	p := color.Palette{color.Black, color.White}
	first := image.NewPaletted(image.Rect(0, 0, 2, 1), p)
	middle := image.NewPaletted(image.Rect(0, 0, 1, 1), p)
	middle.Pix[0] = 1
	last := image.NewPaletted(image.Rect(1, 0, 2, 1), p)
	var b bytes.Buffer
	_ = gif.EncodeAll(&b, &gif.GIF{Image: []*image.Paletted{first, middle, last}, Delay: []int{1, 1, 1}, Disposal: []byte{gif.DisposalNone, gif.DisposalPrevious, gif.DisposalNone}, Config: image.Config{ColorModel: p, Width: 2, Height: 1}})
	m, e := inspectBytes(b.Bytes())
	if e != nil || m["first_last_equal"] != true {
		t.Fatal("did not compare composited endpoints", m, e)
	}
}
func TestStrictRegistryRequiredFields(t *testing.T) {
	dir := copyRegistry(t)
	p := filepath.Join(dir, "models", "gpt-image-2.json")
	b, _ := os.ReadFile(p)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	delete(m, "priority")
	b, _ = json.Marshal(m)
	_ = os.WriteFile(p, b, 0644)
	if _, e := loadRegistry(dir); e == nil {
		t.Fatal("missing required property accepted")
	}
}

func TestExternalRegistryAutoFallback(t *testing.T) {
	dir := copyRegistry(t)
	if err := os.Remove(filepath.Join(dir, "models", defaultModel+".json")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_DRAW_REGISTRY", dir)
	r, err := loadRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, got := range []string{fallbackImageModel(), listedImageModel(nil), (&imageModelSelector{}).Select(nil, settings{}, "auto", true, time.Second)} {
		if got != r.Models[0].ID {
			t.Fatalf("fallback %q is not registered priority candidate %q", got, r.Models[0].ID)
		}
	}
}

func TestBatchRejectsUnsupportedAndTrailingFields(t *testing.T) {
	isolatedSettings(t)
	for _, line := range []string{`{"prompt":"test","scenario":"ui-mockup"}`, `{"prompt":"test"} {"prompt":"extra"}`} {
		p := filepath.Join(t.TempDir(), "jobs.jsonl")
		if err := os.WriteFile(p, []byte(line), 0600); err != nil {
			t.Fatal(err)
		}
		if err := runBatch([]string{"--input", p, "--dry-run"}); err == nil {
			t.Fatal("invalid batch accepted")
		}
	}
}
