---
title: "Cost Analytics"
description: "Discover the two primary cost analysis workflows: Comparative and Trend analysis."
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** focuses on contextual cost analysis. Instead of just showing raw numbers, it helps you understand how your spending is evolving.

## Comparative Workflow

To get a side-by-side comparison of your spending, use the `cost` subcommand:

```bash
aws-doctor cost
```

This triggers the **Comparative Workflow**, which includes a per-service breakdown (EC2, S3, etc.) to help you identify specific cost drivers.

![Comparative Cost Analytics](/images/demo/basic.gif)

### Fairness in Comparison
Traditional billing comparisons (like Month-over-Month) are often misleading. Comparing the 15th of October to the full month of September will always look like a "saving," even if you are spending more.

**AWS Doctor** solves this by comparing identical time windows:
- **Current Period**: 1st day of the month → Today.
- **Previous Period**: 1st day of the previous month → Identical day last month.

*Example: If today is Oct 15th, it compares Oct 1-15 vs Sept 1-15.*

{{< callout type="warning" >}}
**1st Day of the Month**: This feature is not available on the first day of the month. AWS Cost Explorer requires a minimum 24-hour range where the start date is strictly before the end date.
{{< /callout >}}

---

## 6-Month Trend Analysis

To spot long-term growth patterns or sudden architectural shifts, use the `trend` subcommand:

```bash
aws-doctor trend
```

![6-Month Trend Analysis](/images/demo/trend.gif)

### What it shows:
- A high-fidelity ANSI bar chart in your terminal.
- Monthly total costs for the last 6 completed billing cycles.
- Clear indicators of whether your spend is accelerating or stabilizing.
