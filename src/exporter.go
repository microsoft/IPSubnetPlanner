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

	// Write header
	header := []string{"Name", "VLAN", "Subnet", "Prefix", "Network", "Broadcast", "FirstHost", "LastHost", "UsableHosts", "TotalIPs"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data
	for _, result := range results {
		row := []string{
			result.Name,
			fmt.Sprintf("%d", result.VLAN),
			result.Subnet,
			fmt.Sprintf("%d", result.Prefix),
			result.Network,
			result.Broadcast,
			result.FirstHost,
			result.LastHost,
			fmt.Sprintf("%d", result.UsableHosts),
			fmt.Sprintf("%d", result.TotalIPs),
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
	// Collect headers and determine dynamic column widths based on data
	headers := []string{"Name", "VLAN", "Subnet", "Prefix", "Network", "Broadcast", "FirstHost", "LastHost", "UsableHosts", "TotalIPs"}
	colWidths := make([]int, len(headers))

	// Initialize with header lengths
	for i, h := range headers {
		colWidths[i] = len(h)
	}

	// Grow widths based on data content
	for _, r := range results {
		values := []string{
			r.Name,
			fmt.Sprintf("%d", r.VLAN),
			r.Subnet,
			fmt.Sprintf("%d", r.Prefix),
			r.Network,
			r.Broadcast,
			r.FirstHost,
			r.LastHost,
			fmt.Sprintf("%d", r.UsableHosts),
			fmt.Sprintf("%d", r.TotalIPs),
		}
		for i, v := range values {
			if l := len(v); l > colWidths[i] {
				colWidths[i] = l
			}
		}
	}

	// Add minimal padding to each column for readability
	for i := range colWidths {
		colWidths[i] += 2 // 1 space left, 1 space right visual buffer
	}

	// Helper to build format string per row dynamically
	buildFormat := func() string {
		parts := make([]string, len(colWidths))
		for i := range colWidths {
			// Use - to left align. All fields treated as strings when printing generic rows.
			parts[i] = fmt.Sprintf("%%-%ds", colWidths[i])
		}
		return strings.Join(parts, "") + "\n"
	}

	rowFormat := buildFormat()

	// Print header row
	rowVals := make([]interface{}, len(headers))
	for i, h := range headers {
		rowVals[i] = h
	}
	fmt.Printf(rowFormat, rowVals...)

	// Print separator line
	var sepBuilder strings.Builder
	for i, w := range colWidths {
		// Use width of column minus 1 because of padding spaces; ensure at least header length coverage
		dashes := w
		if dashes < len(headers[i]) { // safety, though shouldn't happen
			dashes = len(headers[i])
		}
		sepBuilder.WriteString(strings.Repeat("-", dashes))
	}
	fmt.Println(sepBuilder.String())

	// Print data rows
	for _, r := range results {
		vals := []interface{}{
			r.Name,
			fmt.Sprintf("%d", r.VLAN),
			r.Subnet,
			fmt.Sprintf("%d", r.Prefix),
			r.Network,
			r.Broadcast,
			r.FirstHost,
			r.LastHost,
			fmt.Sprintf("%d", r.UsableHosts),
			fmt.Sprintf("%d", r.TotalIPs),
		}
		fmt.Printf(rowFormat, vals...)
	}
}
