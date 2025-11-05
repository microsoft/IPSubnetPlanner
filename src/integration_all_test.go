package main

import (
	"encoding/json"
	"os"
	"strings"
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
	// Group results by subnet name and category to handle the detailed output format
	subnetMap := make(map[string]map[string]SubnetResult)
	for _, result := range results {
		if result.Name == "Available" && result.VLAN == 0 {
			continue // Skip the overall available space entry
		}
		if subnetMap[result.Name] == nil {
			subnetMap[result.Name] = make(map[string]SubnetResult)
		}
		subnetMap[result.Name][result.Category] = result
	}

	expectedSubnets := []string{"Management", "Users", "Servers"}
	for _, expectedName := range expectedSubnets {
		categories, exists := subnetMap[expectedName]
		if !exists {
			t.Errorf("Expected subnet %s not found in results", expectedName)
			continue
		}

		// Get the network info from any category (they should have the same VLAN and Subnet)
		var networkInfo SubnetResult
		for _, result := range categories {
			networkInfo = result
			break
		}

		// Verify VLAN assignments and basic subnet info
		switch expectedName {
		case "Management":
			if networkInfo.VLAN != 101 {
				t.Errorf("Management VLAN = %d, want 101", networkInfo.VLAN)
			}
			// Should be /27 (30 usable hosts)
			if networkInfo.Prefix != 27 {
				t.Errorf("Management prefix = %d, want 27", networkInfo.Prefix)
			}
			// Check for available IP space
			if available, ok := categories["Available"]; ok {
				if available.TotalIPs < 30 {
					t.Errorf("Management available IPs = %d, want >= 30", available.TotalIPs)
				}
			}
		case "Users":
			if networkInfo.VLAN != 102 {
				t.Errorf("Users VLAN = %d, want 102", networkInfo.VLAN)
			}
			// Should be /25 (126 usable hosts)
			if networkInfo.Prefix != 25 {
				t.Errorf("Users prefix = %d, want 25", networkInfo.Prefix)
			}
			// Check for available IP space
			if available, ok := categories["Available"]; ok {
				if available.TotalIPs < 100 {
					t.Errorf("Users available IPs = %d, want >= 100", available.TotalIPs)
				}
			}
		case "Servers":
			if networkInfo.VLAN != 103 {
				t.Errorf("Servers VLAN = %d, want 103", networkInfo.VLAN)
			}
			// Should be /27 as specified
			if networkInfo.Prefix != 27 {
				t.Errorf("Servers prefix = %d, want 27", networkInfo.Prefix)
			}
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

	// Verify we have the expected subnets (now with detailed entries)
	expectedSubnets := map[string]bool{
		"Management": false,
		"Storage":    false,
		"Compute":    false,
	}

	// Track available IPs per subnet to verify capacity
	availableIPs := make(map[string]int)

	for _, result := range results {
		if _, exists := expectedSubnets[result.Name]; exists {
			expectedSubnets[result.Name] = true

			// Count available IPs (from Available and Unused categories)
			if result.Category == "Available" || result.Category == "Unused" {
				availableIPs[result.Name] += result.TotalIPs
			}

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

	// Verify Storage has enough available IPs (should accommodate 50 hosts)
	if availableIPs["Storage"] < 50 {
		t.Errorf("Storage available IPs = %d, want >= 50", availableIPs["Storage"])
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

	// Should have 4 networks
	if len(networks) != 4 {
		t.Fatalf("Expected 4 networks, got %d", len(networks))
	}

	// Plan all networks
	results, err := PlanSubnets(networks)
	if err != nil {
		t.Fatalf("PlanSubnets() error = %v", err)
	}

	// Verify we have results from multiple networks
	expectedSubnets := map[string]bool{
		"Production-Web":     false,
		"Production-App":     false,
		"Production-DB":      false,
		"Management-Network": false,
		"Storage-Network":    false,
		"Dev-Environment":    false,
		"Test-Environment":   false,
		"CI-CD":              false,
		"Guest-WiFi":         false,
		"IoT-Devices":        false,
	}

	// Track available IPs per subnet to verify capacity
	availableIPs := make(map[string]int)

	for _, result := range results {
		if _, exists := expectedSubnets[result.Name]; exists {
			expectedSubnets[result.Name] = true

			// Count available IPs (from Available and Unused categories)
			if result.Category == "Available" || result.Category == "Unused" {
				availableIPs[result.Name] += result.TotalIPs
			}

			// Verify key VLAN assignments and sizing
			switch result.Name {
			case "Production-Web":
				if result.VLAN != 100 {
					t.Errorf("Production-Web VLAN = %d, want 100", result.VLAN)
				}
				if result.Prefix != 26 {
					t.Errorf("Production-Web prefix = %d, want 26", result.Prefix)
				}
			case "Management-Network":
				if result.VLAN != 200 {
					t.Errorf("Management-Network VLAN = %d, want 200", result.VLAN)
				}
			case "Storage-Network":
				if result.VLAN != 201 {
					t.Errorf("Storage-Network VLAN = %d, want 201", result.VLAN)
				}
			case "Dev-Environment":
				if result.VLAN != 300 {
					t.Errorf("Dev-Environment VLAN = %d, want 300", result.VLAN)
				}
			case "Guest-WiFi":
				if result.VLAN != 400 {
					t.Errorf("Guest-WiFi VLAN = %d, want 400", result.VLAN)
				}
			}
		}
	}

	// Verify capacity for specific subnets
	if availableIPs["Storage-Network"] < 50 {
		t.Errorf("Storage-Network available IPs = %d, want >= 50", availableIPs["Storage-Network"])
	}
	if availableIPs["Dev-Environment"] < 100 {
		t.Errorf("Dev-Environment available IPs = %d, want >= 100", availableIPs["Dev-Environment"])
	}
	if availableIPs["Guest-WiFi"] < 200 {
		t.Errorf("Guest-WiFi available IPs = %d, want >= 200", availableIPs["Guest-WiFi"])
	}

	// Check that all expected subnets were found
	for name, found := range expectedSubnets {
		if !found {
			t.Errorf("Expected subnet %s not found in results", name)
		}
	}

	// Verify we have subnets from all 4 networks by checking unique subnet prefixes
	networkPrefixes := make(map[string]int)
	for _, result := range results {
		// Skip "Available" entries
		if result.Name == "Available" {
			continue
		}
		// Extract network prefix from subnet CIDR (e.g., "10.1." from "10.1.0.0/26")
		if result.Subnet != "" {
			// Split on '/' to get just the IP part
			parts := strings.Split(result.Subnet, "/")
			if len(parts) > 0 {
				// Split the IP into octets
				octets := strings.Split(parts[0], ".")
				if len(octets) >= 2 {
					// Use first two octets to identify network (e.g., "10.1")
					prefix := octets[0] + "." + octets[1]
					networkPrefixes[prefix]++
				}
			}
		}
	}

	// Verify we got subnets from multiple networks
	if len(networkPrefixes) < 3 {
		t.Errorf("Expected subnets from at least 3 different networks, got %d", len(networkPrefixes))
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
	// With the new detailed format, we need to find the Network entry for Large
	for _, result := range results {
		if result.Name == "Large" && result.Category == "Network" {
			// Should start at the beginning of the network
			if result.IP != "10.0.0.0" {
				t.Errorf("Large subnet should start at 10.0.0.0, got %s", result.IP)
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

	// Count available IPs for each subnet
	availableIPs := make(map[string]int)
	for _, result := range results {
		if result.Category == "Available" || result.Category == "Unused" {
			availableIPs[result.Name] += result.TotalIPs
		}
	}

	// Verify capacity
	if availableIPs["P2P"] < 2 {
		t.Errorf("P2P subnet should have at least 2 usable hosts, got %d", availableIPs["P2P"])
	}
	if availableIPs["Host"] < 1 {
		t.Errorf("Host subnet should have at least 1 usable host, got %d", availableIPs["Host"])
	}
	if availableIPs["Normal"] < 5 {
		t.Errorf("Normal subnet should have at least 5 usable hosts, got %d", availableIPs["Normal"])
	}
}
