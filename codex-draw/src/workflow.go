package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var version = "1.0.0"

func runWorkflow(command string, argv []string) error {
	r, e := loadRegistry(registryPath())
	if e != nil {
		return e
	}
	if command == "catalog" || command == "validate" {
		if len(argv) != 0 {
			return errors.New("catalog/validate accept no arguments; use CODEX_DRAW_REGISTRY for another registry")
		}
		if command == "catalog" {
			printJSON(r)
		} else {
			printJSON(map[string]any{"valid": true, "schema_version": 1, "models": len(r.Models), "scenarios": len(r.Scenarios)})
		}
		return nil
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	var id, brief, briefFile, out, mask string
	var images stringList
	f.StringVar(&id, "scenario", "", "scenario ID from catalog")
	f.StringVar(&brief, "brief", "", "task and constraints")
	f.StringVar(&briefFile, "brief-file", "", "UTF-8 brief file")
	f.StringVar(&out, "out", "", "output path")
	f.StringVar(&mask, "mask", "", "PNG alpha mask for edit")
	f.Var(&images, "image", "reference image, repeatable")
	var args commonArgs
	if e := parseCommon(f, argv, &args); e != nil {
		return e
	}
	if command == "plan" {
		args.dryRun = true
	}
	var scenario *scenarioProfile
	for i := range r.Scenarios {
		if r.Scenarios[i].ID == id {
			scenario = &r.Scenarios[i]
			break
		}
	}
	if scenario == nil {
		return fmt.Errorf("unknown scenario %q; run catalog", id)
	}
	text, e := promptValue(brief, briefFile)
	if e != nil {
		return e
	}
	prompt := strings.ReplaceAll(scenario.Prompt, "{{brief}}", text)
	if scenario.ReferenceRequired && len(images) == 0 {
		return errors.New("this scenario requires --image")
	}
	for _, p := range append(append([]string{}, images...), mask) {
		if p == "" {
			continue
		}
		st, e := os.Stat(p)
		if e != nil || st.IsDir() {
			return fmt.Errorf("input file not found: %s", p)
		}
	}
	operation := scenario.Operation
	if len(images) > 0 && operation == "generate" {
		operation = "edit"
	}
	if operation == "edit" && len(images) == 0 {
		return errors.New("edit requires --image")
	}
	if mask != "" && operation != "edit" {
		return errors.New("mask requires edit operation and reference image")
	}
	if args.outputFormat == "" {
		args.outputFormat = scenario.DefaultFormat
	}
	plan := map[string]any{"scenario": id, "prompt": prompt, "operation": operation, "checklist": scenario.Checklist, "inputs_to_review": scenario.Inputs, "format": args.outputFormat, "images": images, "mask": mask, "dry_run": args.dryRun, "availability_checked": false}
	if operation == "handoff" {
		plan["status"] = "handoff_required"
		plan["reason"] = "This adapter produces raster images only. Send this design brief to an SVG/code tool; no generation request was made."
		printJSON(plan)
		if command == "draw" && !args.dryRun {
			return errors.New("SVG generation is not implemented; plan contains the handoff brief")
		}
		return nil
	}
	if e := validateCommon(args); e != nil {
		return e
	}
	requested := args.model
	if requested == "auto" || requested == "" {
		candidates := []modelProfile{}
		for _, m := range r.Models {
			if contains(m.Operations, operation) && contains(m.Formats, args.outputFormat) && (!contains(m.ExperimentalFormats, args.outputFormat) || args.allowExperimental) {
				candidates = append(candidates, m)
			}
		}
		if len(candidates) == 0 {
			return errors.New("no compatible model; experimental formats require --allow-experimental")
		}
		args.model = candidates[0].ID
		if !args.dryRun {
			config, e := resolveSettings(true)
			if e != nil {
				return e
			}
			ids, status, err := fetchModelList(context.Background(), config.APIURL, config.APIKey, 8*time.Second)
			if err == nil && status >= 200 && status < 300 {
				plan["availability_checked"] = true
				for _, m := range candidates {
					if contains(ids, m.ID) {
						args.model = m.ID
						break
					}
				}
			}
		}
	}
	if e := checkModelCapability(args.model, operation, args.outputFormat, args.allowExperimental); e != nil {
		return e
	}
	plan["model"] = args.model
	plan["model_selection"] = requested
	plan["requested_size"] = args.size
	plan["outputs"] = outputPaths(out, args.outDir, prompt, args.n, args.outputFormat)
	if args.dryRun {
		printJSON(plan)
		return nil
	}
	var result map[string]any
	if operation == "edit" {
		result, e = edit(prompt, images, mask, out, args)
	} else {
		result, e = generate(prompt, out, args)
	}
	if e != nil {
		return e
	}
	result["scenario"] = id
	result["prompt"] = prompt
	result["checklist"] = scenario.Checklist
	result["visual_review_required"] = true
	printJSON(result)
	return nil
}

// Explicit migration: only the program decrypts and reseals the old secret.
// Never overwrite a codex-draw config or remove the original.
func importImage2() error {
	store, e := defaultConfigStore()
	if e != nil {
		return e
	}
	if _, e = os.Stat(store.path); !os.IsNotExist(e) {
		return errors.New("target config exists or is inaccessible; import will not overwrite it")
	}
	dir, e := os.UserConfigDir()
	if e != nil {
		return e
	}
	legacy := &configStore{path: filepath.Join(dir, "codex-image2", "config.json"), vault: legacyVault{}}
	if _, e = os.Stat(legacy.path); e != nil {
		return errors.New("no readable legacy config; run configure")
	}
	settings, e := legacy.Resolve(true)
	if e != nil {
		return e
	}
	if e = store.Save(settings.APIURL, settings.APIKey); e != nil {
		return e
	}
	printJSON(map[string]any{"imported": true, "source": "codex-image2", "original_preserved": true})
	return nil
}

type legacyVault struct{}

func (legacyVault) Seal(string) (storedSecret, error) {
	return storedSecret{}, errors.New("legacy vault is read-only")
}
func (legacyVault) Delete(storedSecret)                 {}
func (legacyVault) Open(s storedSecret) (string, error) { return openLegacySecret(s) }
