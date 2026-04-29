---
title: "Waste Detection"
description: "Scan your AWS account for 'zombie' resources. Learn about the detection categories for compute, storage, and networking waste."
weight: 40
type: docs
prev: /docs/cost-analytics
next: /docs/reporting
sidebar:
  collapsed: false
---

The **Waste Detection** engine is the core diagnostic module of **AWS Doctor**. It scans your account for "zombie" resources—assets that are active and billing but provide zero value to your business.

## How to Run
Use the `waste` subcommand to trigger a full scan across all supported services:

```bash
aws-doctor waste --region us-east-1
```

![Waste Detection Scan](/images/demo/waste.gif)

### Selective Scanning
If you only want to scan specific AWS services, you can pass them as arguments to the subcommand. This is useful for faster execution or targeted cleanups.

| Argument | Service |
| :--- | :--- |
| `ec2` | EC2 instances, EBS volumes, snapshots, key pairs, AMIs. |
| `elb` | Application and Network Load Balancers. |
| `s3` | S3 buckets and multipart uploads. |
| `rds` | RDS instances and snapshots. |
| `lambda` | Lambda over-provisioned memory detection. |
| `vpc` | NAT Gateways and idle VPC resources. |
| `cloudwatch` | CloudWatch Log Groups without retention policies. |
| `sagemaker` | SageMaker idle endpoint detection (zero invocations in 14 days). |
| `ecr` | ECR repositories without lifecycle policies, empty repositories, and untagged images. |

```bash
# Example: Scan only EC2 and SageMaker resources
aws-doctor waste ec2 sagemaker
```

### Configuration Flags

The `waste` and `report waste` subcommands support specific flags to tune the detection logic:

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--lambda-memory-threshold` | `10` | Memory utilization threshold (%) below which Lambda functions are flagged as over-provisioned. |

## Categories of Detection

We group waste into primary infrastructure categories:

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="server"
    title="Compute & EBS"
    link="compute/"
    subtitle="EC2 instances stopped for >30 days, orphaned volumes, stale snapshots, and expired RIs."
  >}}
  {{< hextra/feature-card
    icon="code"
    title="Lambda"
    link="compute/#aws-lambda"
    subtitle="Over-provisioned memory detection and right-sizing recommendations."
  >}}
  {{< hextra/feature-card
    icon="database"
    title="Databases"
    link="databases/"
    subtitle="Stopped RDS instances, manual snapshots older than 30 days, and idle database instances."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="Storage & Logs"
    link="storage/"
    subtitle="Buckets without lifecycle policies, hidden incomplete multipart uploads, Log Groups with no retention, and ECR container image waste."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Networking"
    link="networking/"
    subtitle="Unassociated Elastic IPs, idle Load Balancers, and idle NAT Gateways."
  >}}
  {{< hextra/feature-card
    icon="chip"
    title="Machine Learning"
    link="machine-learning/"
    subtitle="Active SageMaker endpoints with zero invocations in the last 14 days."
  >}}
{{< /hextra/feature-grid >}}

---

## Why automate this?
In large organizations, developers often create temporary resources (testing an AMI, spinning up a sandbox EIP) and forget to delete them. Over time, these small charges aggregate into thousands of dollars of "infrastructure debt."

**AWS Doctor** makes it trivial to run a weekly checkup and keep your account lean.
