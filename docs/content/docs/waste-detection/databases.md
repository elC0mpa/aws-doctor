---
title: "Databases (RDS)"
description: "Audit RDS instances and snapshots for waste. Identify idle databases and old manual snapshots to reduce monthly costs."
weight: 15
---

Audit your RDS footprint to eliminate costs from abandoned or underutilized databases.

{{< callout type="info" >}}
**Permissions Required**: `rds:DescribeDBInstances`, `rds:DescribeDBSnapshots`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## RDS Instances

### Stopped Instances
**AWS Doctor** identifies RDS instances that are in a `stopped` state.
- **Reason**: While compute costs are paused, you are still billed for the **allocated storage** and any associated IOPS or multi-AZ configuration.
- **Action**: If the instance is no longer needed, consider deleting it after taking a final snapshot.

### Idle RDS Instances
Finds instances that are in an `available` state but have had **zero database connections** for the last **7 days**.
- **Reason**: These instances are fully operational and billing for compute and storage but are not being utilized by any application.
- **Action**: Stop the instance if it's only needed occasionally, or terminate it if it's no longer used.

{{< callout type="tip" >}}
You can customize the 7-day idle threshold using the `--rds-idle-days` flag.
{{< /callout >}}


---

## RDS Snapshots

### Old Manual Snapshots
Identifies manual RDS snapshots that are **older than 30 days**.
- **Reason**: Unlike automated snapshots which are deleted according to a retention policy, manual snapshots persist indefinitely until manually deleted, incurring ongoing storage costs.
- **Action**: Review old snapshots and delete those that are no longer required for compliance or disaster recovery.

{{< callout type="tip" >}}
You can customize the 30-day age threshold using the `--rds-snapshot-days` flag.
{{< /callout >}}

