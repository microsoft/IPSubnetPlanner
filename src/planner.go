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

		// Handle IP assignments if specified
		if len(req.subnet.IPAssignments) > 0 {
			assignmentResults := processIPAssignments(req.subnet, subnetCIDR, req.prefix)
			results = append(results, assignmentResults...)
		} else {
			// For subnets without IP assignments, create basic entries
			basicResults := createBasicSubnetEntries(req.subnet, subnetCIDR, req.prefix)
			results = append(results, basicResults...)
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
	var results []SubnetResult

	_, ipNet, _ := net.ParseCIDR(cidr)
	networkIP := ipNet.IP.Mask(ipNet.Mask)
	networkInt := ipToUint32(networkIP)

	// Calculate subnet mask
	mask := net.CIDRMask(prefix, 32)

	// Create a map to track assigned positions
	assignedPositions := make(map[int]bool)

	// Add network address entry
	results = append(results, SubnetResult{
		Subnet:   cidr,
		Name:     subnet.Name,
		VLAN:     subnet.VLAN,
		Label:    "Network",
		IP:       networkIP.String(),
		TotalIPs: 1,
		Prefix:   prefix,
		Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
		Category: "Network",
	})

	// Sort assignments by position for consistent ordering
	sort.Slice(subnet.IPAssignments, func(i, j int) bool {
		return subnet.IPAssignments[i].Position < subnet.IPAssignments[j].Position
	})

	// Process IP assignments
	totalIPs := 1 << (32 - prefix)
	for _, assignment := range subnet.IPAssignments {
		var assignedIP net.IP
		position := assignment.Position

		// Handle negative positions (count from end)
		if position < 0 {
			// Negative positions count backwards from broadcast
			if prefix == 32 {
				assignedIP = networkIP
			} else if prefix == 31 {
				assignedIP = uint32ToIP(networkInt + uint32(totalIPs) + uint32(position))
			} else {
				assignedIP = uint32ToIP(networkInt + uint32(totalIPs) - 1 + uint32(position))
			}
		} else if position == 0 {
			// Position 0 means use the network address (for /32 and special cases)
			assignedIP = networkIP
		} else {
			// Positive positions
			assignedIP = uint32ToIP(networkInt + uint32(position))
		}

		assignedPositions[position] = true

		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    assignment.Name,
			IP:       assignedIP.String(),
			TotalIPs: 1,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Assignment",
		})
	}

	// Add unused ranges
	if prefix < 31 {
		// Find gaps in assignments and mark as unused
		usedIPs := make(map[uint32]bool)

		// Mark network address
		usedIPs[networkInt] = true

		// Mark assigned IPs
		for _, assignment := range subnet.IPAssignments {
			position := assignment.Position
			var assignedInt uint32

			if position < 0 {
				if prefix == 31 {
					assignedInt = networkInt + uint32(totalIPs) + uint32(position)
				} else {
					assignedInt = networkInt + uint32(totalIPs) - 1 + uint32(position)
				}
			} else if position == 0 {
				assignedInt = networkInt
			} else {
				assignedInt = networkInt + uint32(position)
			}
			usedIPs[assignedInt] = true
		}

		// Mark broadcast (for non-/31 and non-/32)
		broadcastInt := networkInt + uint32(totalIPs) - 1
		if prefix < 31 {
			usedIPs[broadcastInt] = true
		}

		// Find continuous unused ranges
		rangeStart := -1
		for i := 1; i < totalIPs-1; i++ { // Skip network (0) and broadcast (totalIPs-1)
			currentIP := networkInt + uint32(i)
			if !usedIPs[currentIP] {
				if rangeStart == -1 {
					rangeStart = i
				}
			} else {
				if rangeStart != -1 {
					// End of unused range
					addUnusedRange(&results, subnet, cidr, prefix, mask, networkInt, rangeStart, i-1)
					rangeStart = -1
				}
			}
		}

		// Handle final unused range
		if rangeStart != -1 {
			addUnusedRange(&results, subnet, cidr, prefix, mask, networkInt, rangeStart, totalIPs-2)
		}

		// Add broadcast entry
		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    "Broadcast",
			IP:       uint32ToIP(broadcastInt).String(),
			TotalIPs: 1,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Broadcast",
		})
	}

	return results
}

func addUnusedRange(results *[]SubnetResult, subnet Subnet, cidr string, prefix int, mask net.IPMask, networkInt uint32, start, end int) {
	startIP := uint32ToIP(networkInt + uint32(start))
	endIP := uint32ToIP(networkInt + uint32(end))

	var label string
	count := end - start + 1
	if count == 1 {
		label = "Unused"
	} else {
		label = "Unused Range"
	}

	var ip string
	if count == 1 {
		ip = startIP.String()
	} else {
		ip = fmt.Sprintf("%s - %s", startIP.String(), endIP.String())
	}

	*results = append(*results, SubnetResult{
		Subnet:   cidr,
		Name:     subnet.Name,
		VLAN:     subnet.VLAN,
		Label:    label,
		IP:       ip,
		TotalIPs: count,
		Prefix:   prefix,
		Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
		Category: "Unused",
	})
}

func calculateAvailableSpace(start, end uint32, parentPrefix int) []SubnetResult {
	var results []SubnetResult

	current := start
	for current < end {
		// Find the largest power-of-2 block that fits
		remainingSize := end - current

		// Find largest power of 2 that fits and is aligned
		blockSize := uint32(1)
		maxBlockSize := remainingSize

		// Ensure alignment - block must start at multiple of its size
		for blockSize <= maxBlockSize && (current%blockSize == 0) {
			if blockSize*2 <= maxBlockSize && (current%(blockSize*2) == 0) {
				blockSize *= 2
			} else {
				break
			}
		}

		// Calculate prefix for this block
		prefix := 32 - int(math.Log2(float64(blockSize)))
		if prefix > 32 {
			prefix = 32
		}

		// Calculate actual usable addresses in this block
		usableCount := int(blockSize)
		if prefix < 31 {
			usableCount -= 2 // subtract network and broadcast
		}
		if usableCount < 0 {
			usableCount = 0
		}

		startIP := uint32ToIP(current)
		var label, ip string

		if blockSize == 1 {
			label = "Available"
			ip = startIP.String()
		} else {
			label = "Available Range"
			endIP := uint32ToIP(current + blockSize - 1)
			if prefix < 31 {
				// Show usable range (exclude network and broadcast)
				firstUsable := uint32ToIP(current + 1)
				lastUsable := uint32ToIP(current + blockSize - 2)
				ip = fmt.Sprintf("%s - %s", firstUsable.String(), lastUsable.String())
			} else {
				ip = fmt.Sprintf("%s - %s", startIP.String(), endIP.String())
			}
		}

		// Calculate subnet mask
		mask := net.CIDRMask(prefix, 32)

		result := SubnetResult{
			Subnet:   fmt.Sprintf("%s/%d", startIP.String(), prefix),
			Name:     "Available",
			VLAN:     0,
			Label:    label,
			IP:       ip,
			TotalIPs: usableCount,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Available",
		}
		results = append(results, result)

		current += blockSize
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

func createBasicSubnetEntries(subnet Subnet, cidr string, prefix int) []SubnetResult {
	var results []SubnetResult

	_, ipNet, _ := net.ParseCIDR(cidr)
	networkIP := ipNet.IP.Mask(ipNet.Mask)
	networkInt := ipToUint32(networkIP)
	totalIPs := 1 << (32 - prefix)

	// Calculate subnet mask
	mask := net.CIDRMask(prefix, 32)

	// Add network address entry
	results = append(results, SubnetResult{
		Subnet:   cidr,
		Name:     subnet.Name,
		VLAN:     subnet.VLAN,
		Label:    "Network",
		IP:       networkIP.String(),
		TotalIPs: 1,
		Prefix:   prefix,
		Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
		Category: "Network",
	})

	// Add usable range for normal subnets
	if prefix < 31 {
		firstUsable := uint32ToIP(networkInt + 1)
		lastUsable := uint32ToIP(networkInt + uint32(totalIPs) - 2)
		usableCount := totalIPs - 2

		var label, ip string
		if usableCount == 1 {
			label = "Available"
			ip = firstUsable.String()
		} else {
			label = "Available Range"
			ip = fmt.Sprintf("%s - %s", firstUsable.String(), lastUsable.String())
		}

		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    label,
			IP:       ip,
			TotalIPs: usableCount,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Available",
		})

		// Add broadcast entry
		broadcastIP := uint32ToIP(networkInt + uint32(totalIPs) - 1)
		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    "Broadcast",
			IP:       broadcastIP.String(),
			TotalIPs: 1,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Broadcast",
		})
	} else if prefix == 31 {
		// /31 networks have two usable addresses
		firstIP := networkIP
		secondIP := uint32ToIP(networkInt + 1)

		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    "Available Range",
			IP:       fmt.Sprintf("%s - %s", firstIP.String(), secondIP.String()),
			TotalIPs: 2,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Available",
		})
	} else {
		// /32 networks have one usable address
		results = append(results, SubnetResult{
			Subnet:   cidr,
			Name:     subnet.Name,
			VLAN:     subnet.VLAN,
			Label:    "Available",
			IP:       networkIP.String(),
			TotalIPs: 1,
			Prefix:   prefix,
			Mask:     fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3]),
			Category: "Available",
		})
	}

	return results
}
