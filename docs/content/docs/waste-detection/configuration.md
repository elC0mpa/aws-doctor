---
title: "Configuration and Secrets"
description: "Detect unused AWS Secrets Manager secrets that are billing but no longer accessed."
weight: 70
type: docs
prev: /docs/waste-detection/machine-learning
next: /docs/reporting
---

Secrets stored in **AWS Secrets Manager** cost **$0.40 per secret per month**, regardless of whether they are actively used. Over time, secrets created for decommissioned applications, rotated credentials that were never cleaned up, or one-off testing can accumulate into meaningful recurring charges.

{{< callout type="info" >}}
**Permissions Required**: `secretsmanager:ListSecrets`.
{{< /callout >}}

**AWS Doctor** identifies secrets that have not been accessed within a configurable time window so you can clean them up.

## Secrets Manager

### Detection Logic
A secret is flagged as **unused** if it meets either of the following criteria:
1. **Never accessed**: The `LastAccessedDate` field is null — the secret has never been retrieved since creation.
2. **Stale**: The `LastAccessedDate` is older than the configured idle threshold (default: **90 days**).

{{< callout type="tip" >}}
You can tune the idle threshold using the `--secrets-idle-days` flag (e.g., `--secrets-idle-days 60` to flag secrets not accessed in the last 60 days).
{{< /callout >}}

{{< callout type="info" >}}
Replica secrets (where `PrimaryRegion` differs from the current region) are automatically skipped to avoid double-counting in multi-region setups.
{{< /callout >}}

## How to Run
To run the Secrets Manager waste detection individually:

```bash
aws-doctor waste secrets-manager
```

## Remediation
If a secret is flagged as unused:
1. **Verify**: Confirm the secret is no longer referenced by any application, Lambda function, or CI/CD pipeline.
2. **Delete**: Remove the secret via the AWS Console or CLI. Secrets Manager supports a recovery window (default 30 days) so deletion is reversible.
3. **Rotate**: If the secret is still needed but was simply forgotten, consider enabling automatic rotation to keep it current.
