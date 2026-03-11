---
title: "Waste Detection"
description: "Scan your AWS account for 'zombie' resources. Learn about the detection categories for compute, storage, and networking waste."
weight: 40
type: docs
prev: /docs/cost-analytics
next: /docs/automation
sidebar:
  collapsed: false
---

The **Waste Detection** engine is the core diagnostic module of **AWS Doctor**. It scans your account for "zombie" resources—assets that are active and billing but provide zero value to your business.

## How to Run
Use the `--waste` flag to trigger a full scan across all supported services:

```bash
aws-doctor --waste --region us-east-1
```

![Waste Detection Scan](/images/demo/waste.gif)

### Selective Scanning
If you only want to scan specific AWS services, you can pass a comma-separated list of services directly to the flag. This is useful for faster execution or targeted cleanups. 

Currently supported filters are `ec2`, `s3`, `elb`, and `cloudwatch`.

```bash
# Example: Scan only EC2 and S3 resources
aws-doctor --waste ec2,s3 --region us-east-1

# Example: Scan CloudWatch Logs
aws-doctor --waste cloudwatch --region us-east-1
```

## Categories of Detection

We group waste into three primary infrastructure categories:

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="server"
    title="Compute & EBS"
    link="compute/"
    subtitle="Instances stopped for >30 days, orphaned volumes, stale snapshots, and expired RIs."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="Storage & Logs"
    link="storage/"
    subtitle="Buckets without lifecycle policies, hidden incomplete multipart uploads, and Log Groups with no retention."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Networking"
    link="networking/"
    subtitle="Unassociated Elastic IPs and Load Balancers with no healthy targets."
  >}}
{{< /hextra/feature-grid >}}

---

## Why automate this?
In large organizations, developers often create temporary resources (testing an AMI, spinning up a sandbox EIP) and forget to delete them. Over time, these small charges aggregate into thousands of dollars of "infrastructure debt."

**AWS Doctor** makes it trivial to run a weekly checkup and keep your account lean.
