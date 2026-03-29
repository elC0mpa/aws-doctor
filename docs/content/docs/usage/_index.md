---
title: "Usage Guide"
description: "Detailed guide on using AWS Doctor subcommands and flags, selecting regions and profiles, and managing MFA-protected roles."
weight: 20
type: docs
prev: /docs/getting-started
next: /docs/cost-analytics
---

Learn how to control **AWS Doctor** using subcommands, flags, and configuration profiles.

## Subcommands

These are the primary workflows of the tool.

| Subcommand | Description |
| :--- | :--- |
| `cost` | Run comparative cost analytics (Current month vs. Last month). |
| `waste` | Run the waste detection engine. |
| `trend` | Generate a 6-month cost trend report (optionally filtered by service). |
| `report` | Generate professional PDF reports for cost, waste, or trends. |
| `update` | Self-update the tool to the latest version. |
| `version` | Display version and build information. |
| `help` | Display help for any subcommand. |

## Global Flags

These flags can be used with any subcommand (including the default cost analysis).

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--region` | `~/.aws/config` | Override the target AWS region. |
| `--profile` | `default` | Specify which AWS profile to use. |
| `--output` | `table` | Output format: `table`, `json`, or `csv`. |

---

## Target Selection

### Region Selection
If the `--region` flag is not provided, the tool attempts to find a region in this order:
1. `AWS_REGION` environment variable.
2. `AWS_DEFAULT_REGION` environment variable.
3. The `region` field in your active profile inside `~/.aws/config`.

### Profile Configuration
To run audits against a specific account or role defined in your AWS config:

```bash
aws-doctor cost --profile prod-account
```

---

## MFA Support

**AWS Doctor** has first-class support for Multi-Factor Authentication. If your profile uses `assume_role` with an `mfa_serial`, the tool will detect it and prompt you for your token code securely in the terminal.

```text
Enter MFA code for arn:aws:iam::123456789012:mfa/user: ******
```

{{< callout type="info" >}}
The assumed role session is managed by the tool. You don't need to manually run `aws sts get-session-token`.
{{< /callout >}}

---

## Automatic Updates

Keep your diagnostic engine up to date with a single command:

```bash
aws-doctor update
```
This will check GitHub for the latest release, download the binary for your platform, and replace the existing one.
