---
title: "Storage & Logs"
description: "Optimize S3 storage and CloudWatch Logs costs by identifying buckets without lifecycle policies and Log Groups with no retention."
weight: 20
---

Optimize your S3 and CloudWatch costs by ensuring proper data lifecycle management and cleaning up hidden waste.

{{< callout type="info" >}}
**Permissions Required**: `s3:ListAllMyBuckets`, `s3:GetLifecycleConfiguration`, `s3:ListBucketMultipartUploads`, `logs:DescribeLogGroups`.
{{< /callout >}}

## S3 Lifecycle Policy Audit

**AWS Doctor** scans every bucket in your account to check for an active **Lifecycle Configuration**.

### Why it matters
Without a lifecycle policy, data remains in the (most expensive) Standard storage tier forever unless manually moved. A policy can automate:
- Transitioning old logs to **IA** (Infrequent Access) or **Glacier**.
- Automatically deleting temporary scratch data.
- Deleting old versions of objects.

{{< callout type="warning" >}}
Buckets without lifecycle policies represent a "cost floor" that will only grow over time.
{{< /callout >}}

---

## S3 Incomplete Multipart Uploads

Identifies buckets that have abandoned multipart uploads.

### What are Multipart Uploads?
When you upload a large file to S3, it's broken into parts. If the upload is interrupted or fails, those parts remain in the bucket hidden from the standard object view.

### The Problem
- **Hidden Billing**: You are charged for the storage used by these incomplete parts.
- **Invisibility**: They don't show up in `ls` or standard console views.

**AWS Doctor** counts these hidden parts so you can take action.

### Solution
Add a lifecycle rule to your bucket to **"AbortIncompleteMultipartUpload"** after 7 days.

---

## CloudWatch Log Retention

**AWS Doctor** scans all your **CloudWatch Log Groups** to identify those with no retention policy set (**"Never Expire"**).

### Why it matters
By default, CloudWatch Log Groups are created with an indefinite retention period.
- **Compounding Costs**: As your application runs, logs accumulate, and your monthly bill grows linearly.
- **Storage Debt**: Many developers create log groups for temporary services or debugging and never set a cleanup policy.

### The Problem
- **Cost Leakage**: You pay for every GB stored in CloudWatch. Old, irrelevant logs can account for a significant portion of your monthly AWS bill.
- **Compliance**: Keeping logs indefinitely might violate data privacy regulations like GDPR or SOC2.

### Solution
Set a **Retention Period** (e.g., 30, 90, or 365 days) for each Log Group based on your business and compliance needs.
