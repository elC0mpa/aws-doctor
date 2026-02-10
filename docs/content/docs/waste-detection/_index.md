---
title: "Waste Detection"
weight: 40
type: docs
prev: /docs/cost-analytics
next: /docs/automation
sidebar:
  collapsed: false
---

The **Waste Detection** engine is the core diagnostic module of **AWS Doctor**. It scans your account for "zombie" resources—assets that are active and billing but provide zero value to your business.

## How to Run
Use the `--waste` flag to trigger a multi-service scan:

```bash
aws-doctor --waste --region us-east-1
```

![Waste Detection Scan](/images/demo/waste.gif)

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
    title="Storage"
    link="storage/"
    subtitle="Buckets without lifecycle policies and hidden incomplete multipart uploads."
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
