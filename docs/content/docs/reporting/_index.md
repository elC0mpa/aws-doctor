---
title: "PDF Reporting"
description: "Generate professional, brandable PDF reports for stakeholders with one command. Export cost, waste, and trend findings into a clean document."
weight: 50
type: docs
prev: /docs/waste-detection
next: /docs/automation
---

**AWS Doctor** now supports professional PDF reporting. This feature allows you to transform technical findings into clear, ready-to-share documents for stakeholders and management.

## How to Run

Use the `report` subcommand followed by the type of analysis you want to export.

### 1. Cost Comparison Report
Generate a PDF report comparing the current month's spending against the previous month.

```bash
aws-doctor report cost
```
{{< pdf "/reports/cost.pdf" >}}


### 2. Waste Detection Report
Export all identified "zombie" resources (EC2, S3, RDS, etc.) into a categorized PDF document.

```bash
aws-doctor report waste
```
{{< pdf "/reports/waste.pdf" >}}

### 3. Cost Trend Report
Generate a 6-month visual cost trend report with bar charts and acceleration indicators.

```bash
aws-doctor report trend
```
{{< pdf "/reports/trend.pdf" >}}

---

## Global Flags for Reports

The `report` subcommand supports the standard global flags (`--profile`, `--region`) and a specific flag for the output location.

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--path` | `Documents` | Specify the output directory or full file path for the PDF. |

### Custom Output Path
By default, reports are saved in your **Documents** folder. In case the Documents folder does not exist (common in some Linux distributions), it will be stored in your **Home** folder.

Files are named using the pattern `aws-doctor-<type>-<timestamp>.pdf`. You can override this using the `--path` flag:

```bash
# Save to a specific directory
aws-doctor report cost --path ./my-audits/

# Save to a specific filename
aws-doctor report waste --path ./weekly-scan.pdf
```

---

## Features of the PDF Report

- **Professional Layout**: Clean, minimalist design optimized for readability.
- **Brandable**: Uses the AWS Doctor logo and professional typography.
- **Actionable Summaries**: Each section includes clear totals and identified savings.
- **Charts and Graphs**: Visual representations of cost trends and resource distribution.
- **Categorized Findings**: Resources are grouped by service (Compute, Storage, Networking, etc.) just like in the terminal output.

{{< callout type="info" >}}
**AWS Doctor** generates the PDF entirely locally. Your infrastructure data never leaves your machine during the report generation process.
{{< /callout >}}
