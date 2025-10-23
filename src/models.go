package main

// Network represents a parent network to be subdivided
type Network struct {
	Network string   `json:"network"`
	Subnets []Subnet `json:"subnets"`
}

// Subnet represents a subnet requirement
type Subnet struct {
	Name          string         `json:"name"`
	VLAN          int            `json:"vlan,omitempty"`
	Hosts         int            `json:"hosts,omitempty"`
	CIDR          int            `json:"cidr,omitempty"`
	IPAssignments []IPAssignment `json:"IPAssignments,omitempty"`
}

// IPAssignment represents a named IP address assignment
type IPAssignment struct {
	Name     string `json:"Name"`
	Position int    `json:"Position"`
}

// SubnetResult represents the calculated subnet information
type SubnetResult struct {
	Name        string `json:"name"`
	VLAN        int    `json:"vlan,omitempty"`
	Subnet      string `json:"subnet"`
	Prefix      int    `json:"prefix"`
	Network     string `json:"network"`
	Broadcast   string `json:"broadcast,omitempty"`
	FirstHost   string `json:"firstHost,omitempty"`
	LastHost    string `json:"lastHost,omitempty"`
	UsableHosts int    `json:"usableHosts"`
	TotalIPs    int    `json:"totalIPs"`
	Label       string `json:"label,omitempty"`
	IP          string `json:"ip,omitempty"`
	Mask        string `json:"mask,omitempty"`
	Category    string `json:"category,omitempty"`
}
