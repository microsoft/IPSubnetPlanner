package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sort"
)

// PlanSubnets calculates subnet allocation for a given network
func PlanSubnets(networks []Network) ([]SubnetResult, error) {
	var allResults []SubnetResult

	for _, network := range networks {
		results, err := planSingleNetwork(network)
		if err != nil {
			return nil, fmt.Errorf("error planning network %s: %v", network.Network, err)
		}
		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func planSingleNetwork(network Network) ([]SubnetResult, error) {
	// Parse parent network
	if network.Network == "" {
		return nil, fmt.Errorf("missing 'network' field - each network must specify a CIDR (e.g., \"network\": \"10.0.0.0/24\")")
	}

	_, ipNet, err := net.ParseCIDR(network.Network)
	if err != nil {
		return nil, fmt.Errorf("invalid network CIDR '%s': %v", network.Network, err)
	}

	parentPrefix, _ := ipNet.Mask.Size()
	networkIP := ipNet.IP.Mask(ipNet.Mask)
	networkInt := ipToUint32(networkIP)

	// Calculate required prefix for each subnet
	type subnetReq struct {
		subnet Subnet
		prefix int
		size   uint32
	}

	var requirements []subnetReq
	for _, subnet := range network.Subnets {
		var prefix int
		if subnet.CIDR > 0 {
			prefix = subnet.CIDR
		} else if subnet.Hosts > 0 {
			prefix = calculatePrefixFromHosts(subnet.Hosts)
		} else {
			return nil, fmt.Errorf("subnet %s must specify either 'hosts' or 'cidr'", subnet.Name)
		}

		if prefix < parentPrefix || prefix > 32 {
			return nil, fmt.Errorf("subnet %s: prefix /%d is invalid for parent network /%d", subnet.Name, prefix, parentPrefix)
		}

		size := uint32(1 << (32 - prefix))
		requirements = append(requirements, subnetReq{subnet: subnet, prefix: prefix, size: size})
	}

	// Sort by size (largest first) for optimal allocation
	sort.Slice(requirements, func(i, j int) bool {
		return requirements[i].size > requirements[j].size
	})

	// Allocate subnets
	var results []SubnetResult
	currentIP := networkInt

	for _, req := range requirements {
		subnetIP := uint32ToIP(currentIP)
		subnetCIDR := fmt.Sprintf("%s/%d", subnetIP.String(), req.prefix)

		// Calculate subnet details
		result := calculateSubnetDetails(req.subnet.Name, req.subnet.VLAN, subnetCIDR, req.prefix)

		// Handle IP assignments if specified
		if len(req.subnet.IPAssignments) > 0 {
			assignmentResults := processIPAssignments(req.subnet, subnetCIDR, req.prefix)
			results = append(results, assignmentResults...)
		} else {
			results = append(results, result)
		}

		currentIP += req.size
	}

	// Calculate remaining available space
	parentSize := uint32(1 << (32 - parentPrefix))
	parentEnd := networkInt + parentSize
	if currentIP < parentEnd {
		available := calculateAvailableSpace(currentIP, parentEnd, parentPrefix)
		results = append(results, available...)
	}

	return results, nil
}

func calculatePrefixFromHosts(hosts int) int {
	// Need hosts + 2 (network and broadcast)
	requiredIPs := hosts + 2
	bits := int(math.Ceil(math.Log2(float64(requiredIPs))))
	prefix := 32 - bits
	if prefix < 1 {
		prefix = 1
	}
	if prefix > 30 {
		prefix = 30
	}
	return prefix
}

func calculateSubnetDetails(name string, vlan int, cidr string, prefix int) SubnetResult {
	_, ipNet, _ := net.ParseCIDR(cidr)
	networkIP := ipNet.IP.Mask(ipNet.Mask)

	totalIPs := 1 << (32 - prefix)

	// Special cases for /31 and /32
	if prefix == 31 {
		return SubnetResult{
			Name:        name,
			VLAN:        vlan,
			Subnet:      cidr,
			Prefix:      prefix,
			Network:     networkIP.String(),
			FirstHost:   networkIP.String(),
			LastHost:    uint32ToIP(ipToUint32(networkIP) + 1).String(),
			UsableHosts: 2,
			TotalIPs:    totalIPs,
		}
	}

	if prefix == 32 {
		return SubnetResult{
			Name:        name,
			VLAN:        vlan,
			Subnet:      cidr,
			Prefix:      prefix,
			Network:     networkIP.String(),
			UsableHosts: 1,
			TotalIPs:    totalIPs,
		}
	}

	// Normal subnets
	networkInt := ipToUint32(networkIP)
	broadcast := uint32ToIP(networkInt + uint32(totalIPs) - 1)
	firstHost := uint32ToIP(networkInt + 1)
	lastHost := uint32ToIP(networkInt + uint32(totalIPs) - 2)

	return SubnetResult{
		Name:        name,
		VLAN:        vlan,
		Subnet:      cidr,
		Prefix:      prefix,
		Network:     networkIP.String(),
		Broadcast:   broadcast.String(),
		FirstHost:   firstHost.String(),
		LastHost:    lastHost.String(),
		UsableHosts: totalIPs - 2,
		TotalIPs:    totalIPs,
	}
}

func processIPAssignments(subnet Subnet, cidr string, prefix int) []SubnetResult {
	// TODO: Implement IP assignment logic
	// This would handle the IPAssignments array and create detailed results
	// For now, return basic subnet info
	return []SubnetResult{calculateSubnetDetails(subnet.Name, subnet.VLAN, cidr, prefix)}
}

func calculateAvailableSpace(start, end uint32, parentPrefix int) []SubnetResult {
	var results []SubnetResult

	// Simple implementation - mark remaining space as available
	if start < end {
		remainingSize := end - start
		prefix := 32 - int(math.Log2(float64(remainingSize)))
		cidr := fmt.Sprintf("%s/%d", uint32ToIP(start).String(), prefix)

		result := SubnetResult{
			Name:     "Available",
			Subnet:   cidr,
			Prefix:   prefix,
			Network:  uint32ToIP(start).String(),
			TotalIPs: int(remainingSize),
		}
		results = append(results, result)
	}

	return results
}

// Helper functions
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}
