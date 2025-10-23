package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const version = "1.0.0"

func main() {
	// Define subcommands
	planCmd := flag.NewFlagSet("plan", flag.ExitOnError)
	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Plan command flags
	configFile := planCmd.String("f", "", "Path to JSON configuration file")
	network := planCmd.String("network", "", "Network in CIDR notation (e.g., 192.168.1.0/24)")
	hosts := planCmd.String("hosts", "", "Host requirements (format: 50:2,10:3 means 2 subnets with 50 hosts, 3 with 10)")
	cidr := planCmd.String("cidr", "", "CIDR prefix requirements (format: 26:2,28:3 means 2 /26 subnets, 3 /28)")
	jsonOutput := planCmd.String("json", "", "Export to JSON file")
	csvOutput := planCmd.String("csv", "", "Export to CSV file")
	mdOutput := planCmd.String("md", "", "Export to Markdown file")

	// Check for subcommand
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "plan":
		planCmd.Parse(os.Args[2:])
		handlePlanCommand(*configFile, *network, *hosts, *cidr, *jsonOutput, *csvOutput, *mdOutput)
	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Printf("IPSubnetPlanner version %s\n", version)
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("IPSubnetPlanner - Automated IP subnet planning tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ipsubnetplanner plan [flags]")
	fmt.Println("  ipsubnetplanner version")
	fmt.Println("  ipsubnetplanner help")
	fmt.Println()
	fmt.Println("Plan Flags:")
	fmt.Println("  -f string        Path to JSON configuration file")
	fmt.Println("  -network string  Network in CIDR notation (e.g., 192.168.1.0/24)")
	fmt.Println("  -hosts string    Host requirements (format: 50:2,10:3)")
	fmt.Println("  -cidr string     CIDR prefix requirements (format: 26:2,28:3)")
	fmt.Println("  -json string     Export to JSON file")
	fmt.Println("  -csv string      Export to CSV file")
	fmt.Println("  -md string       Export to Markdown file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # From JSON config")
	fmt.Println("  ipsubnetplanner plan -f network.json")
	fmt.Println()
	fmt.Println("  # Quick planning by hosts")
	fmt.Println("  ipsubnetplanner plan -network 192.168.1.0/24 -hosts 50:2,10:3")
	fmt.Println()
	fmt.Println("  # With exports")
	fmt.Println("  ipsubnetplanner plan -f network.json -json out.json -csv out.csv -md out.md")
}

func handlePlanCommand(configFile, network, hosts, cidr, jsonOutput, csvOutput, mdOutput string) {
	var networks []Network

	// Load from config file
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}

		// Try parsing as array first
		var networkArray []Network
		if err := json.Unmarshal(data, &networkArray); err == nil {
			networks = networkArray
		} else {
			// Try parsing as single network
			var singleNetwork Network
			if err := json.Unmarshal(data, &singleNetwork); err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", err)
				os.Exit(1)
			}
			networks = []Network{singleNetwork}
		}
	} else if network != "" {
		// Quick planning from command line
		// TODO: Implement parsing of hosts/cidr flags
		fmt.Fprintln(os.Stderr, "Quick planning from command line not yet implemented")
		fmt.Fprintln(os.Stderr, "Please use -f flag with a JSON configuration file")
		os.Exit(1)
	} else {
		fmt.Fprintln(os.Stderr, "Error: Either -f or -network flag is required")
		printUsage()
		os.Exit(1)
	}

	// Plan subnets
	results, err := PlanSubnets(networks)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error planning subnets: %v\n", err)
		os.Exit(1)
	}

	// Print results to console
	PrintTable(results)

	// Export if requested
	if jsonOutput != "" {
		ensureDir(jsonOutput)
		if err := ExportJSON(results, jsonOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting JSON: %v\n", err)
		} else {
			fmt.Printf("\n✓ Exported to JSON: %s\n", jsonOutput)
		}
	}

	if csvOutput != "" {
		ensureDir(csvOutput)
		if err := ExportCSV(results, csvOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting CSV: %v\n", err)
		} else {
			fmt.Printf("✓ Exported to CSV: %s\n", csvOutput)
		}
	}

	if mdOutput != "" {
		ensureDir(mdOutput)
		if err := ExportMarkdown(results, mdOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting Markdown: %v\n", err)
		} else {
			fmt.Printf("✓ Exported to Markdown: %s\n", mdOutput)
		}
	}
}

func ensureDir(filePath string) {
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		os.MkdirAll(dir, 0755)
	}
}
