---
title: "Compute & EBS"
description: "Audit EC2 instances, EBS volumes, and snapshots for waste. Identify long-stopped instances and orphaned storage to save costs."
weight: 10
---

Audit your EC2 and EBS footprint to eliminate costs from abandoned instances and data.

{{< callout type="info" >}}
**Permissions Required**: `ec2:DescribeInstances`, `ec2:DescribeReservedInstances`, `ec2:DescribeVolumes`, `ec2:DescribeSnapshots`, `ec2:DescribeKeyPairs`, `ec2:DescribeImages`, `cloudwatch:GetMetricStatistics`, `lambda:ListFunctions`, `logs:DescribeLogGroups`, `logs:StartQuery`, `logs:GetQueryResults`.
{{< /callout >}}

## EC2 Instances

### Long-Stopped Instances
**AWS Doctor** identifies instances that have been in a `stopped` state for **more than 30 days**.
- **Reason**: While you don't pay for CPU/RAM when stopped, you are still paying for the attached EBS root volumes and any persistent storage.
- **Action**: Terminate or snapshot the data and delete.

{{< callout type="tip" >}}
You can customize the 30-day threshold using the `--ec2-stopped-days` flag.
{{< /callout >}}


### Expiring Reserved Instances (RI)
Scans for active RIs scheduled to expire in the **next 30 days** or that have expired in the **last 30 days**.
- **Reason**: Expired RIs revert to expensive On-Demand pricing without warning.
- **Action**: Review usage and renew or migrate to Savings Plans.

{{< callout type="tip" >}}
You can customize the 30-day warning threshold using the `--ec2-ri-expiring-days` flag.
{{< /callout >}}


### Idle Running Instances
Finds `running` instances whose average CPU utilization stayed under **5%** and whose combined NetworkIn + NetworkOut averaged under **5 MB/day** over the last **14 days**.
- **Reason**: Forgotten dev boxes, abandoned workers, and over-sized workloads keep billing for compute, storage, and any attached EIPs while delivering no value.
- **Action**: Stop the instance for a few days to verify nothing notices, then resize to a smaller type or terminate it entirely.

{{< callout type="tip" >}}
You can adjust the sensitivity of this check using the `--ec2-idle-days`, `--ec2-idle-cpu-percent`, and `--ec2-idle-network-bytes` flags.
{{< /callout >}}


---

## AWS Lambda

### Over-Provisioned Memory
Scans for Lambda functions where peak memory utilization is significantly lower than the configured allocation (default threshold: **10%**).
- **Reason**: Lambda pricing is directly proportional to allocated memory. Allocating 10GB to a function that uses 200MB wastes ~98% of the cost.
- **Action**: Right-size the function memory based on the recommendations.
- **Recommendation Engine**: Suggests setting memory to **2x the observed peak** (with a minimum of 128 MB).

{{< callout type="tip" >}}
You can tune the sensitivity of this check using the `--lambda-memory-threshold` flag (e.g., `--lambda-memory-threshold 20` to flag functions using less than 20%). You can also adjust the analysis window using the `--lambda-lookback-days` flag.
{{< /callout >}}


---

## EBS Volumes & Snapshots

### Unused EBS Volumes
Finds volumes with a status of `available` (meaning they are not attached to any instance).
- **Reason**: You are billed for the provisioned size of these volumes every hour they exist.
- **Action**: Delete if no longer needed.

### Orphaned Snapshots
Finds snapshots where the **source volume has been deleted** and the snapshot is not associated with any AMI.
- **Reason**: Often created during manual backups or old deployments and forgotten.
- **Action**: Delete to save on S3-backed storage costs.

### Stale Snapshots & AMIs
Flags AMIs and snapshots that are **older than 90 days** and are not associated with any running or stopped instance.
- **Reason**: Outdated base images and backups that likely haven't been touched in a quarter.
- **Action**: Deregister AMIs and delete snapshots to save storage costs.

{{< callout type="tip" >}}
You can customize the 90-day age thresholds using the `--ec2-ami-stale-days` and `--ec2-snapshot-stale-days` flags.
{{< /callout >}}


---

## Access & Security

### Unused Key Pairs
Identifies EC2 Key Pairs that are not associated with any running or stopped instance.
- **Reason**: Reduces administrative clutter and potential security risks from old keys.
- **Action**: Delete unused keys from the console/CLI.
