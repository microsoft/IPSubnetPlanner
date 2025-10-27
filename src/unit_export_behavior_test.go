package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExportBehavior verifies default markdown export and opt-in JSON/CSV.
func TestExportBehavior(t *testing.T) {
	// Prepare a minimal network to exercise exports directly through core functions
	network := Network{Network: "192.168.1.0/24", Subnets: []Subnet{{Name: "Users", Hosts: 50}}}
	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planning failed: %v", err)
	}

	// Temp directory for exports
	outDir := t.TempDir()

	// 1. Default markdown (simulate main's behavior): Export markdown only
	mdPath := filepath.Join(outDir, "plan.md")
	if err := ExportMarkdown(results, mdPath); err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}
	if _, err := os.Stat(mdPath); err != nil {
		t.Errorf("expected markdown file created: %v", err)
	}

	// 2. Opt-in JSON
	jsonPath := filepath.Join(outDir, "plan.json")
	if err := ExportJSON(results, jsonPath); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Errorf("expected json file created: %v", err)
	}

	// 3. Opt-in CSV
	csvPath := filepath.Join(outDir, "plan.csv")
	if err := ExportCSV(results, csvPath); err != nil {
		t.Fatalf("ExportCSV failed: %v", err)
	}
	if _, err := os.Stat(csvPath); err != nil {
		t.Errorf("expected csv file created: %v", err)
	}
}
