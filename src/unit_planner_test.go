package main

import (
	"net"
	"testing"
)

func TestCalculatePrefixFromHosts(t *testing.T) {
	tests := []struct {
		name     string
		hosts    int
		expected int
	}{
		{"Small network - 5 hosts", 5, 29},
		{"Medium network - 30 hosts", 30, 27},
		{"Large network - 100 hosts", 100, 25},
		{"Very large network - 500 hosts", 500, 23},
		{"Single host", 1, 30},
		{"Two hosts", 2, 30},
		{"Boundary case - 254 hosts", 254, 24},
		{"Boundary case - 1022 hosts", 1022, 22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePrefixFromHosts(tt.hosts)
			if result != tt.expected {
				t.Errorf("calculatePrefixFromHosts(%d) = %d, want %d", tt.hosts, result, tt.expected)
			}
		})
	}
}

func TestIpToUint32AndUint32ToIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want uint32
	}{
		{"Zero IP", "0.0.0.0", 0},
		{"Local IP", "192.168.1.1", 3232235777},
		{"Private IP", "10.0.0.1", 167772161},
		{"Broadcast", "255.255.255.255", 4294967295},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			result := ipToUint32(ip)
			if result != tt.want {
				t.Errorf("ipToUint32(%s) = %d, want %d", tt.ip, result, tt.want)
			}

			// Test reverse conversion
			convertedIP := uint32ToIP(result)
			if !convertedIP.Equal(ip.To4()) {
				t.Errorf("uint32ToIP(%d) = %s, want %s", result, convertedIP.String(), tt.ip)
			}
		})
	}
}

func TestProcessIPAssignments(t *testing.T) {
	tests := []struct {
		name     string
		subnet   Subnet
		cidr     string
		prefix   int
		expected []SubnetResult
	}{
		{
			name: "Simple /28 with IP assignments",
			subnet: Subnet{
				Name: "TestNet",
				VLAN: 100,
				IPAssignments: []IPAssignment{
					{Name: "Gateway", Position: 1},
					{Name: "Server1", Position: 10},
					{Name: "Server2", Position: 11},
				},
			},
			cidr:   "192.168.1.0/28",
			prefix: 28,
			expected: []SubnetResult{
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Network", IP: "192.168.1.0", TotalIPs: 1, Prefix: 28, Mask: "255.255.255.240", Category: "Network"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Gateway", IP: "192.168.1.1", TotalIPs: 1, Prefix: 28, Mask: "255.255.255.240", Category: "Assignment"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Server1", IP: "192.168.1.10", TotalIPs: 1, Prefix: 28, Mask: "255.255.255.240", Category: "Assignment"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Server2", IP: "192.168.1.11", TotalIPs: 1, Prefix: 28, Mask: "255.255.255.240", Category: "Assignment"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Unused Range", IP: "192.168.1.2 - 192.168.1.9", TotalIPs: 8, Prefix: 28, Mask: "255.255.255.240", Category: "Unused"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Unused Range", IP: "192.168.1.12 - 192.168.1.14", TotalIPs: 3, Prefix: 28, Mask: "255.255.255.240", Category: "Unused"},
				{Subnet: "192.168.1.0/28", Name: "TestNet", VLAN: 100, Label: "Broadcast", IP: "192.168.1.15", TotalIPs: 1, Prefix: 28, Mask: "255.255.255.240", Category: "Broadcast"},
			},
		},
		{
			name: "/32 Loopback with position 0",
			subnet: Subnet{
				Name: "Loopback",
				VLAN: 0,
				IPAssignments: []IPAssignment{
					{Name: "Router1", Position: 0},
				},
			},
			cidr:   "192.168.1.1/32",
			prefix: 32,
			expected: []SubnetResult{
				{Subnet: "192.168.1.1/32", Name: "Loopback", VLAN: 0, Label: "Network", IP: "192.168.1.1", TotalIPs: 1, Prefix: 32, Mask: "255.255.255.255", Category: "Network"},
				{Subnet: "192.168.1.1/32", Name: "Loopback", VLAN: 0, Label: "Router1", IP: "192.168.1.1", TotalIPs: 1, Prefix: 32, Mask: "255.255.255.255", Category: "Assignment"},
			},
		},
		{
			name: "/26 with negative position assignments",
			subnet: Subnet{
				Name: "BMC",
				VLAN: 125,
				IPAssignments: []IPAssignment{
					{Name: "Gateway", Position: 1},
					{Name: "BMC", Position: -4},
					{Name: "TOR2", Position: -3},
					{Name: "TOR1", Position: -2},
				},
			},
			cidr:   "10.60.48.128/26",
			prefix: 26,
			expected: []SubnetResult{
				{Subnet: "10.60.48.128/26", Name: "BMC", VLAN: 125, Label: "Network", IP: "10.60.48.128", TotalIPs: 1, Prefix: 26, Mask: "255.255.255.192", Category: "Network"},
				{Subnet: "10.60.48.128/26", Name: "BMC", VLAN: 125, Label: "BMC", IP: "10.60.48.187", TotalIPs: 1, Prefix: 26, Mask: "255.255.255.192", Category: "Assignment"},
				{Subnet: "10.60.48.128/26", Name: "BMC", VLAN: 125, Label: "TOR2", IP: "10.60.48.188", TotalIPs: 1, Prefix: 26, Mask: "255.255.255.192", Category: "Assignment"},
				{Subnet: "10.60.48.128/26", Name: "BMC", VLAN: 125, Label: "TOR1", IP: "10.60.48.189", TotalIPs: 1, Prefix: 26, Mask: "255.255.255.192", Category: "Assignment"},
				{Subnet: "10.60.48.128/26", Name: "BMC", VLAN: 125, Label: "Gateway", IP: "10.60.48.129", TotalIPs: 1, Prefix: 26, Mask: "255.255.255.192", Category: "Assignment"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := processIPAssignments(tt.subnet, tt.cidr, tt.prefix)

			// Check that we have the expected number of results
			if len(results) < len(tt.expected) {
				t.Errorf("Expected at least %d results, got %d", len(tt.expected), len(results))
			}

			// Check for key expected entries
			for _, expected := range tt.expected {
				found := false
				for _, result := range results {
					if result.Label == expected.Label && result.IP == expected.IP {
						found = true
						if result.Subnet != expected.Subnet ||
							result.Name != expected.Name ||
							result.VLAN != expected.VLAN ||
							result.TotalIPs != expected.TotalIPs ||
							result.Category != expected.Category {
							t.Errorf("Result mismatch for %s: got %+v, want %+v", expected.Label, result, expected)
						}
						break
					}
				}
				if !found {
					t.Errorf("Expected result with label %s and IP %s not found", expected.Label, expected.IP)
				}
			}
		})
	}
}

func TestCreateBasicSubnetEntries(t *testing.T) {
	tests := []struct {
		name     string
		subnet   Subnet
		cidr     string
		prefix   int
		expected int // number of expected entries
	}{
		{
			name:     "Standard /28 subnet",
			subnet:   Subnet{Name: "TestNet", VLAN: 100},
			cidr:     "192.168.1.0/28",
			prefix:   28,
			expected: 3, // Network, Available Range, Broadcast
		},
		{
			name:     "/31 point-to-point",
			subnet:   Subnet{Name: "P2P", VLAN: 0},
			cidr:     "192.168.1.0/31",
			prefix:   31,
			expected: 2, // Network, Available Range
		},
		{
			name:     "/32 host route",
			subnet:   Subnet{Name: "Host", VLAN: 0},
			cidr:     "192.168.1.1/32",
			prefix:   32,
			expected: 2, // Network, Available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := createBasicSubnetEntries(tt.subnet, tt.cidr, tt.prefix)

			if len(results) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(results))
			}

			// First entry should always be Network
			if len(results) > 0 && results[0].Category != "Network" {
				t.Errorf("First entry should be Network category, got %s", results[0].Category)
			}

			// For /28 and larger, last entry should be Broadcast
			if tt.prefix < 31 && len(results) > 0 {
				last := results[len(results)-1]
				if last.Category != "Broadcast" {
					t.Errorf("Last entry should be Broadcast category for /%d, got %s", tt.prefix, last.Category)
				}
			}
		})
	}
}

func TestPlanSingleNetwork_SimpleConfig(t *testing.T) {
	network := Network{
		Network: "192.168.1.0/24",
		Subnets: []Subnet{
			{
				Name:  "Management",
				VLAN:  101,
				Hosts: 30,
			},
			{
				Name: "Servers",
				VLAN: 103,
				CIDR: 27,
			},
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one subnet result")
	}

	// Check that we have results for both subnets
	foundManagement := false
	foundServers := false
	for _, result := range results {
		if result.Name == "Management" {
			foundManagement = true
			// Should have network, available, and broadcast entries
			if result.Category == "Network" && result.Prefix != 27 {
				t.Errorf("Management subnet prefix = %d, want 27", result.Prefix)
			}
		}
		if result.Name == "Servers" {
			foundServers = true
			// Servers specified /27 explicitly
			if result.Category == "Network" && result.Prefix != 27 {
				t.Errorf("Servers subnet prefix = %d, want 27", result.Prefix)
			}
		}
	}

	if !foundManagement {
		t.Error("Management subnet not found in results")
	}
	if !foundServers {
		t.Error("Servers subnet not found in results")
	}
}

func TestPlanSingleNetwork_InvalidNetwork(t *testing.T) {
	tests := []struct {
		name    string
		network Network
		wantErr bool
	}{
		{
			name: "Invalid CIDR",
			network: Network{
				Network: "invalid-cidr",
				Subnets: []Subnet{
					{Name: "Test", Hosts: 10},
				},
			},
			wantErr: true,
		},
		{
			name: "Subnet without hosts or CIDR",
			network: Network{
				Network: "192.168.1.0/24",
				Subnets: []Subnet{
					{Name: "Invalid"},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid prefix larger than parent",
			network: Network{
				Network: "192.168.1.0/24",
				Subnets: []Subnet{
					{Name: "TooSmall", CIDR: 16}, // /16 is larger than parent /24
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := planSingleNetwork(tt.network)
			if (err != nil) != tt.wantErr {
				t.Errorf("planSingleNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlanSubnets_MultipleNetworks(t *testing.T) {
	networks := []Network{
		{
			Network: "192.168.1.0/24",
			Subnets: []Subnet{
				{Name: "Compute-VLAN203", VLAN: 203, CIDR: 28},
				{Name: "Compute-VLAN102", VLAN: 102, CIDR: 27},
			},
		},
		{
			Network: "10.50.1.0/24",
			Subnets: []Subnet{
				{Name: "Management-VLAN101", VLAN: 101, CIDR: 27},
				{Name: "Management-VLAN201", VLAN: 201, CIDR: 27},
			},
		},
	}

	results, err := PlanSubnets(networks)
	if err != nil {
		t.Fatalf("PlanSubnets() error = %v", err)
	}

	// Should have at least 4 subnet results (2 per network)
	if len(results) < 4 {
		t.Errorf("Expected at least 4 results, got %d", len(results))
	}

	// Check that all subnet names are present
	expectedNames := []string{
		"Compute-VLAN203", "Compute-VLAN102",
		"Management-VLAN101", "Management-VLAN201",
	}

	foundNames := make(map[string]bool)
	for _, result := range results {
		foundNames[result.Name] = true
	}

	for _, expectedName := range expectedNames {
		if !foundNames[expectedName] {
			t.Errorf("Expected subnet %s not found in results", expectedName)
		}
	}
}

func TestPlanSingleNetwork_WithIPAssignments(t *testing.T) {
	network := Network{
		Network: "192.168.100.0/27",
		Subnets: []Subnet{
			{
				Name: "DMZ-LoadBalancer",
				VLAN: 500,
				CIDR: 28,
				IPAssignments: []IPAssignment{
					{Name: "Firewall", Position: 1},
					{Name: "Reserved", Position: 2},
					{Name: "LoadBalancer", Position: 3},
				},
			},
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Should have Network, Assignment entries, Unused, Broadcast, and Available space
	if len(results) < 5 {
		t.Errorf("Expected at least 5 entries, got %d", len(results))
	}

	// Check for specific IP assignments
	expectedAssignments := map[string]string{
		"Network":      "192.168.100.0",
		"Firewall":     "192.168.100.1",
		"Reserved":     "192.168.100.2",
		"LoadBalancer": "192.168.100.3",
		"Broadcast":    "192.168.100.15",
	}

	foundAssignments := make(map[string]bool)
	for _, result := range results {
		if expectedIP, exists := expectedAssignments[result.Label]; exists {
			if result.IP == expectedIP {
				foundAssignments[result.Label] = true
			}
		}
	}

	for label, expectedIP := range expectedAssignments {
		if !foundAssignments[label] {
			t.Errorf("Expected assignment %s with IP %s not found", label, expectedIP)
		}
	}
}

func TestPlanSubnets_OptimalAllocation(t *testing.T) {
	// Test that larger subnets are allocated first (optimal bin packing)
	network := Network{
		Network: "192.168.1.0/24",
		Subnets: []Subnet{
			{Name: "Small", Hosts: 5},   // Should get /29, but allocated last
			{Name: "Large", Hosts: 100}, // Should get /25, allocated first
			{Name: "Medium", Hosts: 20}, // Should get /27, allocated second
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Verify that the largest subnet gets the first IP range
	for _, result := range results {
		if result.Name == "Large" && result.Category == "Network" {
			// Large subnet should start at network base (192.168.1.0)
			if result.IP != "192.168.1.0" {
				t.Errorf("Large subnet should start at 192.168.1.0, got %s", result.IP)
			}
		}
	}
}

func TestPlanSingleNetwork_NegativePositions(t *testing.T) {
	network := Network{
		Network: "10.60.48.128/26",
		Subnets: []Subnet{
			{
				Name: "BMC-Test",
				VLAN: 125,
				CIDR: 26,
				IPAssignments: []IPAssignment{
					{Name: "Gateway", Position: 1},
					{Name: "BMC", Position: -4},
					{Name: "TOR2", Position: -3},
					{Name: "TOR1", Position: -2},
				},
			},
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	// Check negative position assignments
	expectedNegativeAssignments := map[string]string{
		"BMC":  "10.60.48.187", // -4 from broadcast (191-4=187)
		"TOR2": "10.60.48.188", // -3 from broadcast
		"TOR1": "10.60.48.189", // -2 from broadcast
	}

	foundNegative := make(map[string]bool)
	for _, result := range results {
		if expectedIP, exists := expectedNegativeAssignments[result.Label]; exists {
			if result.IP == expectedIP {
				foundNegative[result.Label] = true
			}
		}
	}

	for label, expectedIP := range expectedNegativeAssignments {
		if !foundNegative[label] {
			t.Errorf("Expected negative position assignment %s with IP %s not found", label, expectedIP)
		}
	}
}

// Benchmark tests
func BenchmarkCalculatePrefixFromHosts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calculatePrefixFromHosts(100)
	}
}

func BenchmarkIpToUint32(b *testing.B) {
	ip := net.ParseIP("192.168.1.1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ipToUint32(ip)
	}
}

func BenchmarkPlanSingleNetwork(b *testing.B) {
	network := Network{
		Network: "10.0.0.0/16",
		Subnets: []Subnet{
			{Name: "Subnet1", Hosts: 100},
			{Name: "Subnet2", Hosts: 50},
			{Name: "Subnet3", CIDR: 26},
			{Name: "Subnet4", CIDR: 28, IPAssignments: []IPAssignment{
				{Name: "Gateway", Position: 1},
				{Name: "Server", Position: 10},
			}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := planSingleNetwork(network)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProcessIPAssignments(b *testing.B) {
	subnet := Subnet{
		Name: "TestSubnet",
		VLAN: 100,
		IPAssignments: []IPAssignment{
			{Name: "Gateway", Position: 1},
			{Name: "Server1", Position: 10},
			{Name: "Server2", Position: 11},
			{Name: "BMC", Position: -2},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processIPAssignments(subnet, "192.168.1.0/24", 24)
	}
}
