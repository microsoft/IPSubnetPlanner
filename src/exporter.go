package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ExportJSON exports results to JSON file
func ExportJSON(results []SubnetResult, filepath string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return os.WriteFile(filepath, data, 0644)
}

// ExportCSV exports results to CSV file
func ExportCSV(results []SubnetResult, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header matching expected format
	header := []string{"Subnet", "Name", "Vlan", "Label", "IP", "TotalIPs", "Prefix", "Mask", "Category"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	for _, result := range results {
		row := []string{
			result.Subnet,
			result.Name,
			fmt.Sprintf("%d", result.VLAN),
			result.Label,
			result.IP,
			fmt.Sprintf("%d", result.TotalIPs),
			fmt.Sprintf("/%d", result.Prefix),
			result.Mask,
			result.Category,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	return nil
}

// ExportMarkdown exports results to Markdown table
func ExportMarkdown(results []SubnetResult, filepath string) error {
	var sb strings.Builder

	// Write header
	sb.WriteString("# Subnet Plan\n\n")
	sb.WriteString("| Name | VLAN | Subnet | Prefix | Network | Broadcast | First Host | Last Host | Usable Hosts | Total IPs |\n")
	sb.WriteString("|------|------|--------|--------|---------|-----------|------------|-----------|--------------|----------|\n")

	// Write data
	for _, result := range results {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d | %s | %s | %s | %s | %d | %d |\n",
			result.Name,
			result.VLAN,
			result.Subnet,
			result.Prefix,
			result.Network,
			result.Broadcast,
			result.FirstHost,
			result.LastHost,
			result.UsableHosts,
			result.TotalIPs,
		))
	}

	return os.WriteFile(filepath, []byte(sb.String()), 0644)
}

// PrintTable prints results as a formatted table to console
func PrintTable(results []SubnetResult) {
	if len(results) == 0 {
		fmt.Println("No subnets generated.")
		return
	}

	fmt.Printf("\nGenerated %d subnet entries:\n\n", len(results))

	// Print header matching CSV format
	fmt.Printf("%-20s %-25s %-6s %-20s %-15s %-10s %-8s %-15s\n",
		"Subnet", "Name", "VLAN", "Label", "IP", "TotalIPs", "Prefix", "Category")
	fmt.Printf("%-20s %-25s %-6s %-20s %-15s %-10s %-8s %-15s\n",
		"------", "----", "----", "-----", "--", "--------", "------", "--------")

	// Print all results in the same format as CSV
	for _, result := range results {
		vlanStr := "-"
		if result.VLAN > 0 {
			vlanStr = fmt.Sprintf("%d", result.VLAN)
		}

		// Handle empty/default values
		label := result.Label
		if label == "" {
			switch result.Category {
			case "Network":
				label = "Network"
			case "Available":
				if strings.Contains(result.IP, " - ") {
					label = "Available Range"
				} else {
					label = "Available"
				}
			case "Broadcast":
				label = "Broadcast"
			case "Assignment":
				label = result.Label // Keep original assignment name
			case "Unused":
				if strings.Contains(result.IP, " - ") {
					label = "Unused Range"
				} else {
					label = "Unused"
				}
			default:
				label = result.Category
			}
		}

		fmt.Printf("%-20s %-25s %-6s %-20s %-15s %-10d %-8s %-15s\n",
			result.Subnet,
			truncate(result.Name, 25),
			vlanStr,
			truncate(label, 20),
			truncate(result.IP, 15),
			result.TotalIPs,
			fmt.Sprintf("/%d", result.Prefix),
			result.Category)
	}

	fmt.Printf("\nThis matches the detailed format in export files.\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
