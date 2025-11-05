package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// version can be set at build time with -ldflags "-X main.version=x.y.z"
var version = "1.0.0"

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	os.Exit(1)
}

// parseSpecs converts spec string value:count pairs into Subnet slice.
// Example hosts spec: "50:2,10:3" => two Host subnets (50) and three Host subnets (10).
func parseSpecs(spec string, isHosts bool) ([]Subnet, error) {
	if spec == "" {
		return nil, nil
	}
	parts := strings.Split(spec, ",")
	var out []Subnet
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.Split(p, ":")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid spec segment: %s", p)
		}
		valueStr, countStr := kv[0], kv[1]
		var value, count int
		if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
			return nil, fmt.Errorf("invalid number in spec: %s", valueStr)
		}
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			return nil, fmt.Errorf("invalid count in spec: %s", countStr)
		}
		if value <= 0 || count <= 0 {
			return nil, fmt.Errorf("value and count must be >0: %s", p)
		}
		for i := 0; i < count; i++ {
			if isHosts {
				out = append(out, Subnet{Name: fmt.Sprintf("hosts-%d-%d", value, i+1), Hosts: value})
			} else {
				out = append(out, Subnet{Name: fmt.Sprintf("cidr-%d-%d", value, i+1), CIDR: value})
			}
		}
	}
	return out, nil
}

func main() {
	// Pre-parse validation to give clearer error if user supplies a bare string export flag without value.
	validateBareOutputFlags()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "IPSubnetPlanner\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  ipsubnetplanner -input config.json\n")
		fmt.Fprintf(os.Stderr, "  ipsubnetplanner -input config.json -exportjson plan.json -exportcsv plan.csv\n")
		fmt.Fprintf(os.Stderr, "  ipsubnetplanner -network 192.168.1.0/24 -hosts 50:2,10:3\n")
		fmt.Fprintf(os.Stderr, "  ipsubnetplanner -network 10.0.0.0/16 -cidr 26:2,28:1\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Flags
	inputFile := flag.String("input", "", "Path to JSON configuration file")
	network := flag.String("network", "", "Parent network in CIDR notation (e.g., 192.168.1.0/24)")
	hostSpec := flag.String("hosts", "", "Host requirements spec (e.g., 50:2,10:3 => 2x50-host, 3x10-host)")
	cidrSpec := flag.String("cidr", "", "CIDR prefix spec (e.g., 26:2,28:1 => 2x/26, 1x/28)")
	exportJSON := flag.String("exportjson", "", "Export to JSON file (disabled by default; specify filename to enable)")
	exportCSV := flag.String("exportcsv", "", "Export to CSV file (disabled by default; specify filename to enable)")
	exportMD := flag.String("exportmd", "plan.md", "Export to Markdown file (default plan.md; set empty to disable)")
	showVersion := flag.Bool("version", false, "Print version and exit")

	// Legacy flag support for backward compatibility
	configFile := flag.String("f", "", "Path to JSON configuration file (deprecated: use -input)")
	jsonOutput := flag.String("json", "", "Export to JSON file (deprecated: use -exportjson)")
	csvOutput := flag.String("csv", "", "Export to CSV file (deprecated: use -exportcsv)")
	mdOutput := flag.String("md", "", "Export to Markdown file (deprecated: use -exportmd)")

	// Legacy compatibility: drop leading 'plan'
	if len(os.Args) > 1 && os.Args[1] == "plan" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
	}

	flag.Parse()

	// Handle flag compatibility - prefer new flags, fall back to legacy ones
	var finalInputFile, finalJSONOutput, finalCSVOutput, finalMDOutput string

	if *inputFile != "" {
		finalInputFile = *inputFile
	} else if *configFile != "" {
		finalInputFile = *configFile
	}

	if *exportJSON != "" {
		finalJSONOutput = *exportJSON
	} else if *jsonOutput != "" {
		finalJSONOutput = *jsonOutput
	}

	if *exportCSV != "" {
		finalCSVOutput = *exportCSV
	} else if *csvOutput != "" {
		finalCSVOutput = *csvOutput
	}

	if *exportMD != "" {
		finalMDOutput = *exportMD
	} else if *mdOutput != "" {
		finalMDOutput = *mdOutput
	} else {
		finalMDOutput = "plan.md" // default
	}

	if *showVersion {
		fmt.Println("IPSubnetPlanner version", version)
		return
	}

	var networks []Network

	if finalInputFile != "" {
		data, err := os.ReadFile(finalInputFile)
		if err != nil {
			fatal(fmt.Sprintf("error reading config file: %v", err))
		}
		// Try array first
		var arr []Network
		if err := json.Unmarshal(data, &arr); err == nil {
			networks = arr
		} else {
			var single Network
			if err := json.Unmarshal(data, &single); err != nil {
				// Provide helpful error message
				errMsg := fmt.Sprintf("error parsing config file: %v\n\n", err)
				errMsg += "Common issues:\n"
				errMsg += "  1. Check that 'vlan' and 'cidr' values are integers (not strings)\n"
				errMsg += "     ✗ Bad:  \"vlan\": \"100\", \"cidr\": \"26\"\n"
				errMsg += "     ✓ Good: \"vlan\": 100, \"cidr\": 26\n\n"
				errMsg += "  2. Verify JSON structure:\n"
				errMsg += "     Single network: {\"network\": \"...\", \"subnets\": [...]}\n"
				errMsg += "     Multi-network:  [{\"network\": \"...\", \"subnets\": [...]}, ...]\n\n"
				errMsg += "See examples/ directory for reference."
				fatal(errMsg)
			}
			networks = []Network{single}
		}
	} else if *network != "" {
		// Build network from specs
		hostSubs, err := parseSpecs(*hostSpec, true)
		if err != nil {
			fatal(err.Error())
		}
		cidrSubs, err := parseSpecs(*cidrSpec, false)
		if err != nil {
			fatal(err.Error())
		}
		if len(hostSubs) == 0 && len(cidrSubs) == 0 {
			fatal("provide at least one -hosts or -cidr spec when using -network")
		}
		networks = []Network{{Network: *network, Subnets: append(hostSubs, cidrSubs...)}}
	} else {
		fatal("either -input (or legacy -f) or -network must be provided")
	}

	results, err := PlanSubnets(networks)
	if err != nil {
		fatal(fmt.Sprintf("planning error: %v", err))
	}

	PrintTable(results)

	// Exports
	if finalJSONOutput != "" {
		ensureDir(finalJSONOutput)
		if err := ExportJSON(results, finalJSONOutput); err != nil {
			fmt.Fprintf(os.Stderr, "error exporting JSON: %v\n", err)
		} else {
			fmt.Printf("\n✓ JSON: %s\n", finalJSONOutput)
		}
	}
	if finalCSVOutput != "" {
		ensureDir(finalCSVOutput)
		if err := ExportCSV(results, finalCSVOutput); err != nil {
			fmt.Fprintf(os.Stderr, "error exporting CSV: %v\n", err)
		} else {
			fmt.Printf("✓ CSV: %s\n", finalCSVOutput)
		}
	}
	if finalMDOutput != "" {
		ensureDir(finalMDOutput)
		if err := ExportMarkdown(results, finalMDOutput); err != nil {
			fmt.Fprintf(os.Stderr, "error exporting Markdown: %v\n", err)
		} else {
			fmt.Printf("✓ Markdown: %s\n", finalMDOutput)
		}
	}
}

func ensureDir(filePath string) {
	dir := filepath.Dir(filePath)
	if dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}
}

// validateBareOutputFlags scans os.Args for a bare occurrence of export flags without a value.
// If found, it prints a clear error and exits before flag.Parse() would produce the generic
// "flag needs an argument" message.
func validateBareOutputFlags() {
	if len(os.Args) == 0 {
		return
	}
	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]
		// Support both old and new flag names
		if arg == "-json" || arg == "--json" || arg == "-csv" || arg == "--csv" || arg == "-md" || arg == "--md" ||
			arg == "-exportjson" || arg == "--exportjson" || arg == "-exportcsv" || arg == "--exportcsv" || arg == "-exportmd" || arg == "--exportmd" {
			// If next token missing or starts with '-' then it's bare.
			if i+1 >= len(os.Args) || strings.HasPrefix(os.Args[i+1], "-") {
				// Tailor message: markdown has a default; json/csv are disabled until filename provided.
				if arg == "-md" || arg == "--md" || arg == "-exportmd" || arg == "--exportmd" {
					fmt.Fprintf(os.Stderr, "Error: %s requires a filename (or use %s=\"\" to disable). Default is plan.md if you omit the flag entirely.\n", arg, arg)
					fmt.Fprintf(os.Stderr, "Tip: Just omit %s to get plan.md automatically.\n", arg)
				} else {
					fmt.Fprintf(os.Stderr, "Error: %s requires a filename (e.g. %s output.json). JSON/CSV exports are disabled unless you provide one.\n", arg, arg)
				}
				os.Exit(2)
			}
		}
	}
}
