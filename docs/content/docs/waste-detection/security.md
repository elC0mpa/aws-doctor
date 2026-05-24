---
title: "Identity & Security"
description: "Detect unused IAM users and root accounts without MFA."
weight: 80
type: docs
prev: /docs/waste-detection/configuration
next: /docs/reporting
---

Identity and Access Management (IAM) is the security backbone of your AWS environment. Inactive users and unprotected root accounts pose significant security risks.

{{< callout type="info" >}}
**Permissions Required**: `iam:GetAccountSummary`, `iam:ListUsers`, `iam:ListAccessKeys`, `iam:GetAccessKeyLastUsed`.
{{< /callout >}}

**AWS Doctor** identifies inactive IAM users and checks if your AWS account's root user lacks Multi-Factor Authentication (MFA).

## IAM Waste & Security

### Detection Logic
A resource is flagged as a security risk or waste if it meets the following criteria:

1. **Unprotected Root Account**: The root account does not have virtual MFA enabled.
2. **Inactive IAM Users**: An IAM user whose password has never been used (or is inactive beyond the idle threshold) AND has all access keys either inactive or unused beyond the idle threshold (default: **90 days**).

{{< callout type="tip" >}}
You can tune the idle threshold for IAM users using the `--iam-idle-days` flag (e.g., `--iam-idle-days 60` to flag users inactive for more than 60 days).
{{< /callout >}}

## How to Run
To run the IAM waste detection individually:

```bash
aws-doctor waste iam
```

## Remediation
If an IAM user or root account is flagged:
1. **Root Account**: Log in as the root user immediately and configure a hardware or virtual MFA device.
2. **IAM Users**: Disable console access or delete access keys that are no longer needed. If the user is an employee who left, delete the IAM user completely.
