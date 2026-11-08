package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestGIFFieldsReachGenerationAndEdit(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", t.TempDir())
	}

	gifBytes := syntheticGIF()
	gifData := base64.StdEncoding.EncodeToString(gifBytes)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		check := func(model, format, frames, duration string) {
			t.Helper()
			if model != "gpt-image-2.6-sunburst" || format != "gif" || frames != "9" || duration != "200" {
				t.Errorf("GIF parameters not passed through: %q %q %q %q", model, format, frames, duration)
			}
		}
		switch r.URL.Path {
		case "/v1/images/generations":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["frame_count"] != float64(9) || body["frame_duration_ms"] != float64(200) {
				t.Error("wrong GIF numeric fields")
			}
			check(body["model"].(string), body["output_format"].(string), "9", "200")
		case "/v1/images/edits":
			if err := r.ParseMultipartForm(2 << 20); err != nil {
				t.Error(err)
			} else {
				check(r.FormValue("model"), r.FormValue("output_format"), r.FormValue("frame_count"), r.FormValue("frame_duration_ms"))
			}
		default:
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"b64_json": gifData}}})
	}))
	defer server.Close()
	t.Setenv("CODEX_API_URL", server.URL)
	t.Setenv("CODEX_API_KEY", "synthetic-gif-test-key")

	args := commonArgs{allowExperimental: true, model: "gpt-image-2.6-sunburst", size: "auto", quality: "auto", n: 1, outDir: t.TempDir(), maxAttempts: 1, timeout: time.Second, outputFormat: "gif", frameCount: 9, frameDurationMS: 200}
	result, err := generate("loop", filepath.Join(t.TempDir(), "generated"), args)
	if err != nil || result["output_format"] != "gif" || result["requested_frame_count"] != 9 || result["requested_frame_duration_ms"] != 200 {
		t.Fatalf("generation GIF request failed: result=%v err=%v", result, err)
	}
	if outputs, ok := result["outputs"].([]string); !ok || filepath.Ext(outputs[0]) != ".gif" {
		t.Fatalf("GIF output extension missing: %v", result["outputs"])
	}

	input := filepath.Join(t.TempDir(), "input.png")
	if err := os.WriteFile(input, []byte("reference"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err = edit("loop", []string{input}, "", filepath.Join(t.TempDir(), "edited"), args)
	if err != nil || result["output_format"] != "gif" || result["requested_frame_count"] != 9 || result["requested_frame_duration_ms"] != 200 {
		t.Fatalf("edit GIF request failed: result=%v err=%v", result, err)
	}
	if result["actual"].([]map[string]any)[0]["frames"] != 2 {
		t.Fatal("actual frames must not echo requested 9")
	}
	if requests != 2 {
		t.Fatalf("expected exactly two mock API requests, got %d", requests)
	}
}
