---
title: "S3 Storage"
description: "Optimize S3 storage costs by identifying buckets without lifecycle policies and detecting hidden incomplete multipart uploads."
weight: 20
---

Optimize your S3 costs by ensuring proper data lifecycle management and cleaning up hidden waste.

{{< callout type="info" >}}
**Permissions Required**: `s3:ListAllMyBuckets`, `s3:GetLifecycleConfiguration`, `s3:ListBucketMultipartUploads`.
{{< /callout >}}

## Lifecycle Policy Audit

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

## Incomplete Multipart Uploads

Identifies buckets that have abandoned multipart uploads.

### What are Multipart Uploads?
When you upload a large file to S3, it's broken into parts. If the upload is interrupted or fails, those parts remain in the bucket hidden from the standard object view.

### The Problem
- **Hidden Billing**: You are charged for the storage used by these incomplete parts.
- **Invisibility**: They don't show up in `ls` or standard console views.

**AWS Doctor** counts these hidden parts so you can take action.

### Solution
Add a lifecycle rule to your bucket to **"AbortIncompleteMultipartUpload"** after 7 days.
