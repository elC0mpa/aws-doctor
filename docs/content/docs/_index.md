---
title: "Documentation"
description: "Welcome to the official AWS Doctor documentation. Explore our guides to cost analytics, waste detection, and automation."
sidebar:
  collapsed: false
---

**AWS Doctor** is a terminal-based health check tool for your AWS infrastructure. It provides immediate context on your spending and identifies "zombie" resources that are costing you money.

## Quick Start

```bash
# Get a comparative cost analysis (Current month vs Last month)
aws-doctor cost

# Scan for idle resources and waste
aws-doctor waste

# View a 6-month spending trend
aws-doctor trend
```

## Explore the Modules

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="presentation-chart-line"
    title="Usage Guide"
    link="usage/"
    subtitle="Detailed explanation of CLI subcommands, global flags, MFA support, and profile management."
  >}}
  {{< hextra/feature-card
    icon="calculator"
    title="Cost Analytics"
    link="cost-analytics/"
    subtitle="Understanding the comparative billing engine and the 6-month trend visualization."
  >}}
  {{< hextra/feature-card
    icon="trash"
    title="Waste Detection"
    link="waste-detection/"
    subtitle="Deep dive into zombie resource detection: EC2, S3, ELB, and CloudWatch."
  >}}
  {{< hextra/feature-card
    icon="cpu-chip"
    title="Automation"
    link="automation/"
    subtitle="Using JSON output to integrate AWS Doctor into your CI/CD pipelines."
  >}}
{{< /hextra/feature-grid >}}

---

## Need Help?
Looking for something specific? Use the search bar at the top of the page to find details about a particular service or subcommand.
