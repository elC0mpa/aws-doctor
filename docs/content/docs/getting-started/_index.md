---
title: "Getting Started"
description: "Learn how to install AWS Doctor and configure the necessary AWS credentials and permissions to start auditing your infrastructure."
weight: 10
type: docs
next: /docs/usage
---

Get up and running with **AWS Doctor** in less than a minute.

## Installation

### 1. One-Line Script (Linux & macOS)
The fastest way to install the latest version:

```bash
curl -sSfL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh
```

### 2. Using Go
If you have Go installed (1.23+):

```bash
go install github.com/elC0mpa/aws-doctor@latest
```

### 3. Manual Binary Download
Download the pre-compiled binary for your architecture from the [GitHub Releases](https://github.com/elC0mpa/aws-doctor/releases) page. Supported platforms:
- **macOS** (Intel & Apple Silicon)
- **Linux** (amd64 & arm64)
- **Windows** (amd64)

---

## Prerequisites

### AWS Credentials
**AWS Doctor** uses the standard AWS Go SDK. It will automatically look for credentials in:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`).
2. Shared credentials file (`~/.aws/credentials`).
3. IAM Roles for EC2/ECS if running inside AWS.

### Minimum Permissions
The tool requires **Read-Only** access to perform audits.

{{< callout type="info" >}}
**Zero-Risk Execution**: For the simplest and safest experience, we recommend using the AWS managed policy **`ReadOnlyAccess`**. This ensures the tool has the necessary visibility across all services to execute every flow without any modification capabilities.
{{< /callout >}}

While `ReadOnlyAccess` is the easiest way to get started, **AWS Doctor** also supports granular IAM policies. Each functionality described in this documentation (such as [S3 Storage](../waste-detection/storage/) or [Compute](../waste-detection/compute/)) includes a dedicated section listing the exact IAM permissions required.
