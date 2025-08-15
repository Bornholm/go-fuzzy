package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bornholm/go-fuzzy/dsl"
	"github.com/pkg/errors"
)

// LoadDSLFiles loads all .dsl files from the specified directory
func loadFiles(pattern string) (map[string]string, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, errors.Errorf("failed to find files with pattern '%s': %+v", pattern, err)
	}

	dslFiles := make(map[string]string)
	for _, f := range files {
		// Extract name without extension (my-engine.dsl -> my-engine)
		name := strings.TrimSuffix(filepath.Base(f), filepath.Ext(f))

		content, err := os.ReadFile(f)
		if err != nil {
			return nil, errors.Errorf("failed to read file %s: %+v", f, err)
		}

		dslFiles[name] = string(content)
	}

	return dslFiles, nil
}

// createRegistryFromDSL parses DSL content and creates associated registry
func createRegistryFromDSL(dslFiles map[string]string) (*Registry, error) {
	registry := NewRegistry()

	for name, content := range dslFiles {
		// Parse rules and variables
		result, err := dsl.ParseRulesAndVariables(content)
		if err != nil {
			return nil, errors.Errorf("failed to parse DSL for engine %s: %+v", name, err)
		}

		// Register the engine
		registry.Register(name, result.Variables, result.Rules)
	}

	return registry, nil
}

func main() {
	config := parseConfig()

	// Load DSL files
	log.Printf("Loading fuzzy engine definition files from pattern '%s'", config.Definitions)

	dslFiles, err := loadFiles(config.Definitions)
	if err != nil {
		log.Fatalf("Failed to load dsl files: %v", err)
	}

	if len(dslFiles) == 0 {
		log.Printf("No files found with pattern '%s'", config.Definitions)
	} else {
		// Get engine names and join them for logging
		engineNames := make([]string, 0, len(dslFiles))
		for name := range dslFiles {
			engineNames = append(engineNames, name)
		}
		log.Printf("Loaded %d definition files: %v", len(dslFiles), strings.Join(engineNames, ", "))
	}

	// Create engines from DSL files
	registry, err := createRegistryFromDSL(dslFiles)
	if err != nil {
		log.Fatalf("Failed to create engines: %v", err)
	}

	// Create HTTP handler
	handler := createHandler(registry)

	handler = loggingMiddleware(handler)

	// Start HTTP server
	log.Printf("Starting server on %s", config.Address)
	log.Fatal(http.ListenAndServe(config.Address, handler))
}
