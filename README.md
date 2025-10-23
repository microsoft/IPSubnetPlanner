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
* Export JSON / CSV / Markdown

## Quick Start (Download & Run)
1. Go to Releases: https://github.com/microsoft/IPSubnetPlanner/releases
2. Download the latest binary for your OS (Windows / Linux / macOS)
3. (Linux/macOS) Make it executable: `chmod +x ipsubnetplanner*`
4. Run with the provided example:
  ```bash
  ./ipsubnetplanner plan -f examples/simple.json
  ```
5. (Optional) Export:
  ```bash
  ./ipsubnetplanner plan -f examples/simple.json -json plan.json -csv plan.csv -md plan.md
  ```

Need to customize? Create your own JSON (see use cases below).

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
Run:
```bash
./ipsubnetplanner plan -f single.json
```
Sample Output:
```
Name         VLAN  Subnet             Prefix  FirstHost       LastHost        UsableHosts
Users        102   192.168.1.0/25     25      192.168.1.1     192.168.1.126   126
Management   101   192.168.1.128/27   27      192.168.1.129   192.168.1.158   30
Servers      103   192.168.1.160/27   27      192.168.1.161   192.168.1.190   30
Available          192.168.1.192/26   26      192.168.1.193   192.168.1.254   62
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
Export variants:
```bash
./ipsubnetplanner plan -f advanced.json -json plan.json -csv plan.csv -md plan.md
```
Excerpt (Markdown):
```
| Name       | VLAN | Subnet      | Gateway    | DNS1       | DNS2       | LastHost   |
|------------|------|-------------|------------|------------|------------|------------|
| Management | 100  | 10.0.0.0/28 | 10.0.0.1   | 10.0.0.2   | 10.0.0.3   | 10.0.0.14  |
```

### 3. Multi‑Network Planning
`multi.json`
```json
[
  {
    "network": "192.168.10.0/24",
    "subnets": [
      { "name": "Edge", "cidr": 27 },
      { "name": "Users", "hosts": 90 }
    ]
  },
  {
    "network": "10.50.1.0/24",
    "subnets": [
      { "name": "Mgmt", "cidr": 27 },
      { "name": "Compute", "cidr": 26 }
    ]
  }
]
```
Run:
```bash
./ipsubnetplanner plan -f multi.json
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

## Commands
```bash
ipsubnetplanner plan -f config.json
ipsubnetplanner plan -f config.json -json out.json -csv out.csv -md out.md
ipsubnetplanner help
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