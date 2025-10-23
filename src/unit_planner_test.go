package main

import (
	"net"
	"reflect"
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

func TestCalculateSubnetDetails(t *testing.T) {
	tests := []struct {
		name     string
		subnet   string
		vlan     int
		cidr     string
		prefix   int
		expected SubnetResult
	}{
		{
			name:   "Standard /24 subnet",
			subnet: "Production",
			vlan:   100,
			cidr:   "192.168.1.0/24",
			prefix: 24,
			expected: SubnetResult{
				Name:        "Production",
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
		},
		{
			name:   "Small /28 subnet",
			subnet: "Management",
			vlan:   101,
			cidr:   "192.168.1.0/28",
			prefix: 28,
			expected: SubnetResult{
				Name:        "Management",
				VLAN:        101,
				Subnet:      "192.168.1.0/28",
				Prefix:      28,
				Network:     "192.168.1.0",
				Broadcast:   "192.168.1.15",
				FirstHost:   "192.168.1.1",
				LastHost:    "192.168.1.14",
				UsableHosts: 14,
				TotalIPs:    16,
			},
		},
		{
			name:   "Point-to-point /31 subnet",
			subnet: "P2P-Link",
			vlan:   200,
			cidr:   "192.168.1.0/31",
			prefix: 31,
			expected: SubnetResult{
				Name:        "P2P-Link",
				VLAN:        200,
				Subnet:      "192.168.1.0/31",
				Prefix:      31,
				Network:     "192.168.1.0",
				FirstHost:   "192.168.1.0",
				LastHost:    "192.168.1.1",
				UsableHosts: 2,
				TotalIPs:    2,
			},
		},
		{
			name:   "Host route /32 subnet",
			subnet: "LoopbackIP",
			vlan:   0,
			cidr:   "192.168.1.1/32",
			prefix: 32,
			expected: SubnetResult{
				Name:        "LoopbackIP",
				VLAN:        0,
				Subnet:      "192.168.1.1/32",
				Prefix:      32,
				Network:     "192.168.1.1",
				UsableHosts: 1,
				TotalIPs:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateSubnetDetails(tt.subnet, tt.vlan, tt.cidr, tt.prefix)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("calculateSubnetDetails() = %+v, want %+v", result, tt.expected)
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
			// Management needs 30 hosts, so should get /27 (32 addresses - 2 = 30 usable)
			if result.Prefix != 27 {
				t.Errorf("Management subnet prefix = %d, want 27", result.Prefix)
			}
			if result.UsableHosts != 30 {
				t.Errorf("Management subnet usable hosts = %d, want 30", result.UsableHosts)
			}
		}
		if result.Name == "Servers" {
			foundServers = true
			// Servers specified /27 explicitly
			if result.Prefix != 27 {
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

func TestPlanSubnets_LargeHostRequirements(t *testing.T) {
	network := Network{
		Network: "10.0.0.0/16", // Large network
		Subnets: []Subnet{
			{Name: "Large-Users", Hosts: 1000}, // Should get /22
			{Name: "Small-Mgmt", Hosts: 10},    // Should get /28
		},
	}

	results, err := planSingleNetwork(network)
	if err != nil {
		t.Fatalf("planSingleNetwork() error = %v", err)
	}

	for _, result := range results {
		switch result.Name {
		case "Large-Users":
			if result.Prefix != 22 {
				t.Errorf("Large-Users prefix = %d, want 22 (for 1000+ hosts)", result.Prefix)
			}
			if result.UsableHosts < 1000 {
				t.Errorf("Large-Users usable hosts = %d, want >= 1000", result.UsableHosts)
			}
		case "Small-Mgmt":
			if result.Prefix != 28 {
				t.Errorf("Small-Mgmt prefix = %d, want 28 (for 10+ hosts)", result.Prefix)
			}
			if result.UsableHosts < 10 {
				t.Errorf("Small-Mgmt usable hosts = %d, want >= 10", result.UsableHosts)
			}
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
		if result.Name == "Large" {
			// Large subnet should start at network base (192.168.1.0)
			if result.Network != "192.168.1.0" {
				t.Errorf("Large subnet should start at 192.168.1.0, got %s", result.Network)
			}
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
			{Name: "Subnet4", CIDR: 28},
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