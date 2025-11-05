# IPSubnetPlanner

[![Build](https://github.com/microsoft/IPSubnetPlanner/actions/workflows/build_artifacts.yml/badge.svg)](https://github.com/microsoft/IPSubnetPlanner/actions/workflows/build_artifacts.yml)
[![Tests](https://github.com/microsoft/IPSubnetPlanner/actions/workflows/unit_test.yml/badge.svg)](https://github.com/microsoft/IPSubnetPlanner/actions/workflows/unit_test.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Fast, zero‑friction IP subnet planning from a simple JSON file. No spreadsheets. No manual math.

> Internal utility we found useful and open sourced. Not a supported Microsoft product.

## Features (What You Get)
* Plan subnets by host count OR fixed CIDR
* Optimal packing (largest first) + remaining space summary
* Optional VLAN IDs & named IP assignments
* Multi‑network (array) input
* **Detailed console preview** showing all IP assignments and breakdowns
* Markdown export by default (plan.md)
* Opt-in JSON / CSV (only if you supply filenames)

## Quick Start (Download & Run)
1. Go to Releases: https://github.com/microsoft/IPSubnetPlanner/releases
2. Download the latest binary for your OS (Windows / Linux / macOS)
3. (Linux/macOS) Make it executable: `chmod +x ipsubnetplanner*`
4. Run with the provided example (prints detailed table + writes plan.md):
  ```bash
  ./ipsubnetplanner -input examples/simple.json
  ```
5. (Optional) Opt-in exports (you must give filenames for JSON / CSV):
  ```bash
  ./ipsubnetplanner -input examples/simple.json -exportjson plan.json -exportcsv plan.csv -exportmd report.md
  ```

Need to customize? Create your own JSON (see use cases below).

## Architecture Overview

### High-Level Flow

```mermaid
flowchart TD
    A[User runs ipsubnetplanner] --> B{Input Method}
    B -->|JSON file| C[-input config.json]
    B -->|CLI flags| D[-network -hosts -cidr]
    C --> E[Network Definitions]
    D --> E
    E --> F[Planning Engine]
    F --> G[Subnet Results]
    G --> H[Console Output]
    G --> I[File Exports]
    
    style H fill:#d4edff,stroke:#0078d4
    style I fill:#fff4ce,stroke:#ffd700
```

### Planning Engine Details

```mermaid
flowchart TD
    Start[Network + Subnets] --> Sort[Sort by size<br/>largest first]
    Sort --> Prefix[Calculate<br/>minimal prefix]
    Prefix --> Allocate[Allocate<br/>IP blocks]
    Allocate --> Compute[Compute ranges<br/>network/broadcast/hosts]
    Compute --> VLAN[Apply VLAN<br/>assignments]
    VLAN --> Remaining[Calculate<br/>remaining space]
    Remaining --> Result[Subnet Results]
    
    style Start fill:#e8f4ff,stroke:#4a90e2
    style Result fill:#e0ffe8,stroke:#55aa55
```

### Export Behavior

```mermaid
flowchart LR
    Results[Subnet Results] --> Console[Console Table]
    Results --> MD{Markdown}
    Results --> JSON{JSON}
    Results --> CSV{CSV}
    
    MD -->|Always on| MD1[plan.md ✓]
    MD -->|Can disable<br/>-exportmd=""| MD2[Skipped]
    
    JSON -->|Off by default| JSON1[No file]
    JSON -->|Opt-in<br/>-exportjson file.json| JSON2[file.json ✓]
    
    CSV -->|Off by default| CSV1[No file]
    CSV -->|Opt-in<br/>-exportcsv file.csv| CSV2[file.csv ✓]
    
    style MD1 fill:#e0ffe8,stroke:#55aa55
    style JSON2 fill:#e8f4ff,stroke:#4a90e2
    style CSV2 fill:#fff4e0,stroke:#e28b2c
    style Console fill:#d4edff,stroke:#0078d4
```

**Key Points:**
* **Input flexibility**: JSON file or inline CLI flags (`-network`, `-hosts`, `-cidr`)
* **Smart allocation**: Largest subnets first to minimize fragmentation
* **Detailed console output**: Shows all IP assignments, ranges, and categories in terminal
* **Default export**: Markdown always created unless explicitly disabled
* **Opt-in exports**: JSON/CSV only generated when you specify filenames


---
## Top 3 Use Cases

### 1. Single Network (Mixed Hosts + CIDR)
`single.json`
```json
{
  "network": "192.168.1.0/24",
  "subnets": [
    { "name": "Users", "hosts": 100, "vlan": 102 },
    { "name": "Management", "hosts": 30, "vlan": 101 },
    { "name": "Servers", "cidr": 27, "vlan": 103 }
  ]
}
```
Run (table + plan.md):
```bash
./ipsubnetplanner -input single.json
```
Sample Output:
```
Subnet               Name                      VLAN   Label                IP              TotalIPs   Prefix   Category       
------               ----                      ----   -----                --              --------   ------   --------
192.168.1.0/25       Users                     102    Network              192.168.1.0     1          /25      Network
192.168.1.0/25       Users                     102    Available Range      192.168.1.1-126 126        /25      Available
192.168.1.0/25       Users                     102    Broadcast            192.168.1.127   1          /25      Broadcast
192.168.1.128/27     Management                101    Network              192.168.1.128   1          /27      Network
192.168.1.128/27     Management                101    Available Range      192.168.1.129-158 30       /27      Available
192.168.1.128/27     Management                101    Broadcast            192.168.1.159   1          /27      Broadcast
192.168.1.160/27     Servers                   103    Network              192.168.1.160   1          /27      Network
192.168.1.160/27     Servers                   103    Available Range      192.168.1.161-190 30       /27      Available
192.168.1.160/27     Servers                   103    Broadcast            192.168.1.191   1          /27      Broadcast
192.168.1.192/26     Available                 -      Available Range      192.168.1.193-254 62       /26      Available
```

### 2. VLAN + Named IP Assignments
`advanced.json`
```json
{
  "network": "10.0.0.0/24",
  "subnets": [
    {
      "name": "Management",
      "cidr": 28,
      "vlan": 100,
      "IPAssignments": [
        { "Name": "Gateway", "Position": 1 },
        { "Name": "DNS1", "Position": 2 },
        { "Name": "DNS2", "Position": 3 },
        { "Name": "LastHost", "Position": -1 }
      ]
    },
    { "name": "Storage", "hosts": 50, "vlan": 110 },
    { "name": "Compute", "cidr": 26, "vlan": 120 }
  ]
}
```
Opt-in exports:
```bash
./ipsubnetplanner -input advanced.json -exportjson plan.json -exportcsv plan.csv -exportmd plan.md
```
Console Output shows detailed IP assignments:
```
Subnet               Name                      VLAN   Label                IP              TotalIPs   Prefix   Category       
------               ----                      ----   -----                --              --------   ------   --------
10.0.0.128/28        Management                110    Network              10.0.0.128      1          /28      Network
10.0.0.128/28        Management                110    Gateway              10.0.0.129      1          /28      Assignment
10.0.0.128/28        Management                110    DNS1                 10.0.0.130      1          /28      Assignment
10.0.0.128/28        Management                110    DNS2                 10.0.0.131      1          /28      Assignment
10.0.0.128/28        Management                110    LastHost             10.0.0.142      1          /28      Assignment
```
CSV Export:
```
Subnet,Name,Vlan,Label,IP,TotalIPs,Prefix,Mask,Category
10.0.0.128/28,Management,110,Gateway,10.0.0.129,1,/28,255.255.255.240,Assignment
10.0.0.128/28,Management,110,DNS1,10.0.0.130,1,/28,255.255.255.240,Assignment
```

### 3. Multi‑Network Planning
`multi-network.json`
```json
[
  {
    "network": "10.1.0.0/24",
    "subnets": [
      { "name": "Production-Web", "vlan": 100, "cidr": 26 },
      { "name": "Production-App", "vlan": 101, "cidr": 26 },
      { "name": "Production-DB", "vlan": 102, "cidr": 27 }
    ]
  },
  {
    "network": "10.2.0.0/24",
    "subnets": [
      { "name": "Management-Network", "vlan": 200, "cidr": 26 },
      { "name": "Storage-Network", "vlan": 201, "hosts": 50 }
    ]
  },
  {
    "network": "10.3.0.0/24",
    "subnets": [
      { "name": "Dev-Environment", "vlan": 300, "hosts": 100 },
      { "name": "Test-Environment", "vlan": 301, "hosts": 50 }
    ]
  }
]
```
Run:
```bash
./ipsubnetplanner -input multi-network.json
```

---
## Minimal Config Reference
Subnet (choose hosts OR cidr):
```json
{
  "name": "Web", "hosts": 120, "vlan": 10
}
```
Field | Meaning
------|--------
hosts | Required host count (tool picks smallest fitting prefix)
cidr | Fixed prefix length (1–32)
vlan | Optional VLAN ID (0–4094)
IPAssignments | Array of { Name, Position }

IP Positions:
* 1 = first usable host, 2 = second, etc.
* -1 = last address, -2 = second last
* 0 allowed only when vlan = 0 (special /31 or /32 contexts)

Rules:
* Exactly one of hosts or cidr
* Largest required subnets allocated first
* Remaining space reported as "Available"

## Console Output
The tool displays a detailed table in the terminal showing **exactly the same data** as the export files:
- **Network entries**: Base network address and subnet mask
- **Assignment entries**: Individual IP assignments with their names/purposes
- **Available entries**: IP ranges available for future use
- **Broadcast entries**: Broadcast addresses
- **Unused entries**: Gaps between assignments

This eliminates confusion between terminal preview and export files - what you see is what you get!

## Commands

```bash
ipsubnetplanner -input config.json                          # detailed table + plan.md
ipsubnetplanner -input config.json -exportmd design.md      # override markdown filename
ipsubnetplanner -input config.json -exportmd=""             # disable markdown export
ipsubnetplanner -input config.json -exportjson out.json     # enable JSON export
ipsubnetplanner -input config.json -exportcsv out.csv       # enable CSV export
ipsubnetplanner -input config.json -exportjson out.json -exportcsv out.csv -exportmd report.md
ipsubnetplanner -network 192.168.1.0/24 -hosts 50:2,10:3   # CLI input with host requirements
ipsubnetplanner -network 10.0.0.0/16 -cidr 26:2,28:1       # CLI input with fixed CIDR sizes
ipsubnetplanner -version
```

## Build From Source
```bash
cd IPSubnetPlanner/src
go build -o ../ipsubnetplanner
```
Cross‑compile:
```bash
GOOS=linux GOARCH=amd64   go build -o dist/ipsubnetplanner-linux-amd64 ./src
GOOS=windows GOARCH=amd64 go build -o dist/ipsubnetplanner-windows-amd64.exe ./src
```

## Optional: Run Tests
```bash
cd src
go test -v
go test -cover
```