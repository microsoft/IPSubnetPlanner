package main

import (
	"encoding/json"
	"os"
	"testing"
)

// Integration tests using the example configurations
func TestIntegration_SimpleExample(t *testing.T) {
	// Load simple.json example
	data, err := os.ReadFile("../examples/simple.json")
	if err != nil {
		t.Fatalf("Failed to read simple.json: %v", err)
	}

	var network Network
	err = json.Unmarshal(data, &network)
	if err != nil {
		t.Fatalf("Failed to unmarshal simple.json: %v", err)
	}

	// Plan the network
	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Verify we have the expected subnets
	expectedSubnets := map[string]bool{
		"Management": false,
		"Users":      false,
		"Servers":    false,
	}

	for _, result := range results {
		if _, exists := expectedSubnets[result.Name]; exists {
			expectedSubnets[result.Name] = true

			// Verify VLAN assignments
			switch result.Name {
			case "Management":
				if result.VLAN != 101 {
					t.Errorf("Management VLAN = %d, want 101", result.VLAN)
				}
				// Should accommodate 30 hosts
				if result.UsableHosts < 30 {
					t.Errorf("Management usable hosts = %d, want >= 30", result.UsableHosts)
				}
			case "Users":
				if result.VLAN != 102 {
					t.Errorf("Users VLAN = %d, want 102", result.VLAN)
				}
				// Should accommodate 100 hosts
				if result.UsableHosts < 100 {
					t.Errorf("Users usable hosts = %d, want >= 100", result.UsableHosts)
				}
			case "Servers":
				if result.VLAN != 103 {
					t.Errorf("Servers VLAN = %d, want 103", result.VLAN)
				}
				// Should be /27 as specified
				if result.Prefix != 27 {
					t.Errorf("Servers prefix = %d, want 27", result.Prefix)
				}
			}
		}
	}

	// Check that all expected subnets were found
	for name, found := range expectedSubnets {
		if !found {
			t.Errorf("Expected subnet %s not found in results", name)
		}
	}
}

func TestIntegration_AdvancedExample(t *testing.T) {
	// Load advanced.json example
	data, err := os.ReadFile("../examples/advanced.json")
	if err != nil {
		t.Fatalf("Failed to read advanced.json: %v", err)
	}

	var network Network
	err = json.Unmarshal(data, &network)
	if err != nil {
		t.Fatalf("Failed to unmarshal advanced.json: %v", err)
	}

	// Plan the network
	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Verify we have the expected subnets
	expectedSubnets := map[string]bool{
		"Management": false,
		"Storage":    false,
		"Compute":    false,
	}

	for _, result := range results {
		if _, exists := expectedSubnets[result.Name]; exists {
			expectedSubnets[result.Name] = true

			// Verify specific requirements
			switch result.Name {
			case "Management":
				if result.VLAN != 110 {
					t.Errorf("Management VLAN = %d, want 110", result.VLAN)
				}
				if result.Prefix != 28 {
					t.Errorf("Management prefix = %d, want 28", result.Prefix)
				}
			case "Storage":
				if result.VLAN != 120 {
					t.Errorf("Storage VLAN = %d, want 120", result.VLAN)
				}
				// Should accommodate 50 hosts
				if result.UsableHosts < 50 {
					t.Errorf("Storage usable hosts = %d, want >= 50", result.UsableHosts)
				}
			case "Compute":
				if result.VLAN != 130 {
					t.Errorf("Compute VLAN = %d, want 130", result.VLAN)
				}
				if result.Prefix != 26 {
					t.Errorf("Compute prefix = %d, want 26", result.Prefix)
				}
			}
		}
	}

	// Check that all expected subnets were found
	for name, found := range expectedSubnets {
		if !found {
			t.Errorf("Expected subnet %s not found in results", name)
		}
	}
}

func TestIntegration_MultiNetworkExample(t *testing.T) {
	// Load multi-network.json example
	data, err := os.ReadFile("../examples/multi-network.json")
	if err != nil {
		t.Fatalf("Failed to read multi-network.json: %v", err)
	}

	var networks []Network
	err = json.Unmarshal(data, &networks)
	if err != nil {
		t.Fatalf("Failed to unmarshal multi-network.json: %v", err)
	}

	// Plan all networks
	results, err := PlanSubnets(networks)
	if err != nil {
		t.Fatalf("PlanSubnets() error = %v", err)
	}

	// Verify we have results from both networks
	expectedSubnets := map[string]bool{
		"Compute-VLAN203":    false,
		"Compute-VLAN102":    false,
		"Management-VLAN101": false,
		"Management-VLAN201": false,
	}

	for _, result := range results {
		if _, exists := expectedSubnets[result.Name]; exists {
			expectedSubnets[result.Name] = true

			// Verify VLAN assignments
			switch result.Name {
			case "Compute-VLAN203":
				if result.VLAN != 203 {
					t.Errorf("Compute-VLAN203 VLAN = %d, want 203", result.VLAN)
				}
				if result.Prefix != 28 {
					t.Errorf("Compute-VLAN203 prefix = %d, want 28", result.Prefix)
				}
			case "Compute-VLAN102":
				if result.VLAN != 102 {
					t.Errorf("Compute-VLAN102 VLAN = %d, want 102", result.VLAN)
				}
				if result.Prefix != 27 {
					t.Errorf("Compute-VLAN102 prefix = %d, want 27", result.Prefix)
				}
			case "Management-VLAN101":
				if result.VLAN != 101 {
					t.Errorf("Management-VLAN101 VLAN = %d, want 101", result.VLAN)
				}
				if result.Prefix != 27 {
					t.Errorf("Management-VLAN101 prefix = %d, want 27", result.Prefix)
				}
			case "Management-VLAN201":
				if result.VLAN != 201 {
					t.Errorf("Management-VLAN201 VLAN = %d, want 201", result.VLAN)
				}
				if result.Prefix != 27 {
					t.Errorf("Management-VLAN201 prefix = %d, want 27", result.Prefix)
				}
			}
		}
	}

	// Check that all expected subnets were found
	for name, found := range expectedSubnets {
		if !found {
			t.Errorf("Expected subnet %s not found in results", name)
		}
	}

	// Verify we have subnets from both networks
	var network1Subnets, network2Subnets int
	for _, result := range results {
		// Check IP ranges to determine which network
		if result.Network != "" {
			if result.Network[:7] == "192.168" {
				network1Subnets++
			} else if result.Network[:2] == "10" {
				network2Subnets++
			}
		}
	}

	if network1Subnets == 0 {
		t.Error("No subnets found for 192.168.1.0/24 network")
	}
	if network2Subnets == 0 {
		t.Error("No subnets found for 10.50.1.0/24 network")
	}
}

func TestIntegration_EndToEndWithExports(t *testing.T) {
	// Test complete workflow: load config -> plan -> export
	data, err := os.ReadFile("../examples/simple.json")
	if err != nil {
		t.Fatalf("Failed to read simple.json: %v", err)
	}

	var network Network
	err = json.Unmarshal(data, &network)
	if err != nil {
		t.Fatalf("Failed to unmarshal simple.json: %v", err)
	}

	// Plan the network
	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Test all export formats
	tempDir := t.TempDir()

	// Test JSON export
	jsonFile := tempDir + "/test_export.json"
	err = ExportJSON(results, jsonFile)
	if err != nil {
		t.Errorf("ExportJSON() error = %v", err)
	}

	// Test CSV export
	csvFile := tempDir + "/test_export.csv"
	err = ExportCSV(results, csvFile)
	if err != nil {
		t.Errorf("ExportCSV() error = %v", err)
	}

	// Test Markdown export
	mdFile := tempDir + "/test_export.md"
	err = ExportMarkdown(results, mdFile)
	if err != nil {
		t.Errorf("ExportMarkdown() error = %v", err)
	}

	// Verify all files were created
	for _, file := range []string{jsonFile, csvFile, mdFile} {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Export file %s was not created", file)
		}
	}
}

func TestIntegration_SubnetAllocationOrder(t *testing.T) {
	// Test that larger subnets are allocated first for optimal space usage
	network := Network{
		Network: "10.0.0.0/24",
		Subnets: []Subnet{
			{Name: "Small1", Hosts: 5},  // /29 - 8 IPs
			{Name: "Large", Hosts: 60},  // /26 - 64 IPs
			{Name: "Medium", Hosts: 20}, // /27 - 32 IPs
			{Name: "Small2", Hosts: 10}, // /28 - 16 IPs
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Find the Large subnet and verify it got the first allocation
	for _, result := range results {
		if result.Name == "Large" {
			// Should start at the beginning of the network
			if result.Network != "10.0.0.0" {
				t.Errorf("Large subnet should start at 10.0.0.0, got %s", result.Network)
			}
			// Should have /26 prefix
			if result.Prefix != 26 {
				t.Errorf("Large subnet should have /26 prefix, got /%d", result.Prefix)
			}
			break
		}
	}
}

func TestIntegration_NetworkCapacityValidation(t *testing.T) {
	// Test that we can't allocate more subnets than the parent network can hold
	network := Network{
		Network: "192.168.1.0/29", // Only 8 IPs total
		Subnets: []Subnet{
			{Name: "TooLarge", Hosts: 10}, // Needs more than 8 IPs
		},
	}

	_, err := planSingleNetwork(network)
	if err == nil {
		t.Error("Expected error when subnet requirement exceeds parent network capacity")
	}
}

func TestIntegration_EdgeCaseSubnets(t *testing.T) {
	// Test edge cases: /31 and /32 subnets
	network := Network{
		Network: "192.168.1.0/28", // 16 IPs
		Subnets: []Subnet{
			{Name: "P2P", CIDR: 31},    // Point-to-point
			{Name: "Host", CIDR: 32},   // Single host
			{Name: "Normal", Hosts: 5}, // Normal subnet
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	for _, result := range results {
		switch result.Name {
		case "P2P":
			if result.UsableHosts != 2 {
				t.Errorf("P2P subnet should have 2 usable hosts, got %d", result.UsableHosts)
			}
		case "Host":
			if result.UsableHosts != 1 {
				t.Errorf("Host subnet should have 1 usable host, got %d", result.UsableHosts)
			}
		case "Normal":
			if result.UsableHosts < 5 {
				t.Errorf("Normal subnet should have at least 5 usable hosts, got %d", result.UsableHosts)
			}
		}
	}
}
