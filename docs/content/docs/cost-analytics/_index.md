---
title: "Cost Analytics"
description: "Understand the 'Fair Assessment' logic behind AWS Doctor's cost comparisons and how to generate 6-month trend reports."
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** provides context-aware cost analysis that goes beyond simple totals.

{{< callout type="info" >}}
**Permissions Required**: `ce:GetCostAndUsage`
{{< /callout >}}

## Comparative Cost Analytics

When you run `aws-doctor` without flags, it triggers the **Comparative Workflow**. This includes a per-service breakdown (EC2, S3, etc.) to help you identify specific cost drivers.

![Comparative Cost Analytics](/images/demo/basic.gif)

### The "Fair Assessment" Logic
Most billing tools compare the current month's total against the previous month's total. This is often misleading (e.g., comparing 10 days of spending in February against 31 days in January).

**AWS Doctor** compares identical time windows:
- **Current Period**: 1st day of current month → Today.
- **Previous Period**: 1st day of previous month → Identical day last month.

*Example: If today is October 15th, it compares Oct 1–15 against Sep 1–15.*

{{< callout type="warning" >}}
**1st Day of the Month**: This feature is unavailable on the 1st day of the month. AWS Cost Explorer requires a minimum 24-hour range where the start date is strictly before the end date.
{{< /callout >}}

---

## 6-Month Trend Analysis

To spot long-term growth patterns or sudden architectural shifts, use the `--trend` flag:

```bash
aws-doctor --trend
```

![6-Month Trend Analysis](/images/demo/trend.gif)

### What it shows:
- A high-fidelity ANSI bar chart in your terminal.
- Monthly total costs for the last 6 completed billing cycles.
- Clear indicators of whether your spending is accelerating or stabilizing.
