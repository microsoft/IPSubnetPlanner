package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportJSON(t *testing.T) {
	// Create test data
	testResults := []SubnetResult{
		{
			Name:        "TestSubnet",
			VLAN:        100,
			Subnet:      "192.168.1.0/24",
			Prefix:      24,
			Network:     "192.168.1.0",
			Broadcast:   "192.168.1.255",
			FirstHost:   "192.168.1.1",
			LastHost:    "192.168.1.254",
			UsableHosts: 254,
			TotalIPs:    256,
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_export.json")

	// Test export
	err := ExportJSON(testResults, testFile)
	if err != nil {
		t.Fatalf("ExportJSON() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Exported JSON file does not exist")
	}

	// Read and verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read exported JSON file: %v", err)
	}

	var importedResults []SubnetResult
	err = json.Unmarshal(data, &importedResults)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}

	if len(importedResults) != len(testResults) {
		t.Errorf("Expected %d results, got %d", len(testResults), len(importedResults))
	}

	if importedResults[0].Name != testResults[0].Name {
		t.Errorf("Expected name %s, got %s", testResults[0].Name, importedResults[0].Name)
	}
}

func TestExportCSV(t *testing.T) {
	// Create test data
	testResults := []SubnetResult{
		{
			Subnet:   "192.168.1.0/28",
			Name:     "Subnet1",
			VLAN:     100,
			Label:    "Network",
			IP:       "192.168.1.0",
			TotalIPs: 1,
			Prefix:   28,
			Mask:     "255.255.255.240",
			Category: "Network",
		},
		{
			Subnet:   "192.168.1.0/28",
			Name:     "Subnet1",
			VLAN:     100,
			Label:    "Gateway",
			IP:       "192.168.1.1",
			TotalIPs: 1,
			Prefix:   28,
			Mask:     "255.255.255.240",
			Category: "Assignment",
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_export.csv")

	// Test export
	err := ExportCSV(testResults, testFile)
	if err != nil {
		t.Fatalf("ExportCSV() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Exported CSV file does not exist")
	}

	// Read and verify content
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open exported CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read CSV records: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("Expected 3 CSV records (header + 2 data), got %d", len(records))
	}

	// Check header
	expectedHeader := []string{"Subnet", "Name", "Vlan", "Label", "IP", "TotalIPs", "Prefix", "Mask", "Category"}
	if len(records[0]) != len(expectedHeader) {
		t.Errorf("Expected %d header columns, got %d", len(expectedHeader), len(records[0]))
	}

	// Check first data row
	if records[1][0] != "192.168.1.0/28" {
		t.Errorf("Expected first record subnet '192.168.1.0/28', got '%s'", records[1][0])
	}
	if records[1][1] != "Subnet1" {
		t.Errorf("Expected first record name 'Subnet1', got '%s'", records[1][1])
	}
}

func TestExportMarkdown(t *testing.T) {
	// Create test data
	testResults := []SubnetResult{
		{
			Name:        "Management",
			VLAN:        100,
			Subnet:      "192.168.1.0/28",
			Prefix:      28,
			Network:     "192.168.1.0",
			Broadcast:   "192.168.1.15",
			FirstHost:   "192.168.1.1",
			LastHost:    "192.168.1.14",
			UsableHosts: 14,
			TotalIPs:    16,
		},
	}

	// Create temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_export.md")

	// Test export
	err := ExportMarkdown(testResults, testFile)
	if err != nil {
		t.Fatalf("ExportMarkdown() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("Exported Markdown file does not exist")
	}

	// Read and verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read exported Markdown file: %v", err)
	}

	content := string(data)

	// Check for expected content
	if !strings.Contains(content, "# Subnet Plan") {
		t.Error("Markdown should contain '# Subnet Plan' header")
	}
	if !strings.Contains(content, "| Name | VLAN |") {
		t.Error("Markdown should contain table header")
	}
	if !strings.Contains(content, "| Management |") {
		t.Error("Markdown should contain Management subnet data")
	}
	if !strings.Contains(content, "|------|") {
		t.Error("Markdown should contain table separator")
	}
}

func TestExportJSON_EmptyResults(t *testing.T) {
	// Test with empty results
	var emptyResults []SubnetResult

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "empty_export.json")

	err := ExportJSON(emptyResults, testFile)
	if err != nil {
		t.Fatalf("ExportJSON() with empty results error = %v", err)
	}

	// Read and verify
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read exported JSON file: %v", err)
	}

	var importedResults []SubnetResult
	err = json.Unmarshal(data, &importedResults)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}

	if len(importedResults) != 0 {
		t.Errorf("Expected 0 results, got %d", len(importedResults))
	}
}

func TestExportCSV_InvalidPath(t *testing.T) {
	testResults := []SubnetResult{
		{Name: "Test", Subnet: "192.168.1.0/24"},
	}

	// Try to export to invalid path (assuming /invalid/path doesn't exist)
	err := ExportCSV(testResults, "/invalid/path/test.csv")
	if err == nil {
		t.Error("Expected error when exporting to invalid path, got nil")
	}
}

func TestExportMarkdown_MultipleSubnets(t *testing.T) {
	// Test with multiple subnets to ensure all are included
	testResults := []SubnetResult{
		{
			Name:        "DMZ",
			VLAN:        10,
			Subnet:      "192.168.1.0/26",
			Prefix:      26,
			Network:     "192.168.1.0",
			Broadcast:   "192.168.1.63",
			FirstHost:   "192.168.1.1",
			LastHost:    "192.168.1.62",
			UsableHosts: 62,
			TotalIPs:    64,
		},
		{
			Name:        "LAN",
			VLAN:        20,
			Subnet:      "192.168.1.64/26",
			Prefix:      26,
			Network:     "192.168.1.64",
			Broadcast:   "192.168.1.127",
			FirstHost:   "192.168.1.65",
			LastHost:    "192.168.1.126",
			UsableHosts: 62,
			TotalIPs:    64,
		},
		{
			Name:        "MGMT",
			VLAN:        30,
			Subnet:      "192.168.1.128/27",
			Prefix:      27,
			Network:     "192.168.1.128",
			Broadcast:   "192.168.1.159",
			FirstHost:   "192.168.1.129",
			LastHost:    "192.168.1.158",
			UsableHosts: 30,
			TotalIPs:    32,
		},
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "multi_subnet.md")

	err := ExportMarkdown(testResults, testFile)
	if err != nil {
		t.Fatalf("ExportMarkdown() error = %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read exported Markdown file: %v", err)
	}

	content := string(data)

	// Check that all subnets are present
	expectedSubnets := []string{"DMZ", "LAN", "MGMT"}
	for _, subnet := range expectedSubnets {
		if !strings.Contains(content, subnet) {
			t.Errorf("Markdown should contain subnet '%s'", subnet)
		}
	}

	// Count the number of data rows (should be 3)
	lines := strings.Split(content, "\n")
	dataRows := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "| ") && !strings.Contains(line, "Name | VLAN") && !strings.Contains(line, "---") {
			dataRows++
		}
	}

	if dataRows != 3 {
		t.Errorf("Expected 3 data rows in Markdown table, got %d", dataRows)
	}
}

// Benchmark tests for export functions
func BenchmarkExportJSON(b *testing.B) {
	// Create test data
	testResults := make([]SubnetResult, 100)
	for i := 0; i < 100; i++ {
		testResults[i] = SubnetResult{
			Name:        "Subnet" + string(rune(i)),
			VLAN:        100 + i,
			Subnet:      "192.168.1.0/28",
			Prefix:      28,
			Network:     "192.168.1.0",
			Broadcast:   "192.168.1.15",
			FirstHost:   "192.168.1.1",
			LastHost:    "192.168.1.14",
			UsableHosts: 14,
			TotalIPs:    16,
		}
	}

	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "benchmark.json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExportJSON(testResults, testFile)
	}
}

func BenchmarkExportCSV(b *testing.B) {
	// Create test data
	testResults := make([]SubnetResult, 100)
	for i := 0; i < 100; i++ {
		testResults[i] = SubnetResult{
			Name:        "Subnet" + string(rune(i)),
			VLAN:        100 + i,
			Subnet:      "192.168.1.0/28",
			Prefix:      28,
			Network:     "192.168.1.0",
			Broadcast:   "192.168.1.15",
			FirstHost:   "192.168.1.1",
			LastHost:    "192.168.1.14",
			UsableHosts: 14,
			TotalIPs:    16,
		}
	}

	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "benchmark.csv")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ExportCSV(testResults, testFile)
	}
}
