# aws-doctor

<p align="center">
  <a href="https://github.com/elC0mpa/aws-doctor/blob/main/go.mod"><img src="https://img.shields.io/github/go-mod/go-version/elC0mpa/aws-doctor" alt="Go Version"></a>
  <a href="https://pkg.go.dev/github.com/elC0mpa/aws-doctor"><img src="https://pkg.go.dev/badge/github.com/elC0mpa/aws-doctor.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/elC0mpa/aws-doctor"><img src="https://goreportcard.com/badge/github.com/elC0mpa/aws-doctor" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/elC0mpa/aws-doctor"><img src="https://codecov.io/gh/elC0mpa/aws-doctor/graph/badge.svg" alt="codecov"></a>
</p>

<p align="center">
  <a href="https://github.com/elC0mpa/aws-doctor/releases"><img src="https://img.shields.io/github/downloads/elC0mpa/aws-doctor/total?color=blue&label=Downloads" alt="GitHub all releases"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/actions/workflows/ci.yml"><img src="https://github.com/elC0mpa/aws-doctor/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/elC0mpa/aws-doctor" alt="License"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/commits/main"><img src="https://img.shields.io/badge/Maintained-yes-green.svg" alt="Maintained"></a>
</p>

A terminal-based tool that acts as a comprehensive health check for your AWS accounts. Built with Golang, **aws-doctor** diagnoses cost anomalies, detects idle resources, and provides a proactive analysis of your cloud infrastructure—effectively giving you the insights of AWS Trusted Advisor without the need for a Business or Enterprise support plan.

![](https://github.com/elC0mpa/aws-cost-billing/blob/main/assets/logo.webp)

## Demo

### Basic usage

![](https://github.com/elC0mpa/aws-cost-billing/blob/main/demo/basic.gif)

### Trend

![](https://github.com/elC0mpa/aws-cost-billing/blob/main/demo/trend.gif)

### Waste

![](https://github.com/elC0mpa/aws-cost-billing/blob/main/demo/waste.gif)

## Features

- **📉 Cost Comparison:** Compares costs between the current and previous month for the exact same period (e.g., comparing Jan 1–15 vs Feb 1–15) to give a fair assessment of spending velocity.

> [!IMPORTANT]
> This feature is not available on the **1st day of the month** as AWS Cost Explorer requires a minimum 24-hour range (Start date must be before End date).

- **🏥 Waste Detection (The "Checkup"):** Scans your account for "zombie" resources and inefficiencies that are silently inflating your bill.
- **📊 Trend Analysis:** Visualizes cost history over the last 6 months to spot long-term anomalies.
- **🔐 MFA Support:** Fully supports AWS profiles that require Multi-Factor Authentication (MFA) to assume roles.
- **🎨 Startup Banner:** Displays a high-fidelity ANSI logo on launch. Automatically adjusts colors based on your terminal background (e.g., switches to AmazonOrange on blue backgrounds) and can be customized via the `AWS_DOCTOR_BANNER_COLOR` environment variable.

## Motivation

As a Cloud Architect, I often need to check AWS costs and billing information. While the AWS Console provides raw data, it lacks the immediate context I need to answer the question: _"Are we spending efficiently?"_

I created **aws-doctor** to fill that gap. It doesn't just show you the bill; it acts as a diagnostic tool that helps you understand **where** the money is going and **what** can be cleaned up. It automates the routine checks I used to perform manually, serving as a free, open-source alternative to the paid recommendations found in AWS Trusted Advisor.

## Installation

### Quick Install (macOS/Linux)

```bash
curl -sSfL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh
```

### Using Go

```bash
go install github.com/elC0mpa/aws-doctor@latest
```

### Download Binary

Download the latest release for your platform from the [Releases page](https://github.com/elC0mpa/aws-doctor/releases).

Available platforms:

- macOS (Intel & Apple Silicon)
- Linux (amd64 & arm64)
- Windows (amd64)

> [!TIP]
> Once installed, you can keep **aws-doctor** up to date by running `aws-doctor --update`.

## Flags

- `--profile`: Specify the AWS profile to use. Supports MFA-protected role assumption.
- `--region`: Specify the AWS region to use. If not provided, uses `AWS_REGION` or `AWS_DEFAULT_REGION` environment variables, or the region from `~/.aws/config`.
- `--trend`: Shows a trend analysis for the last 6 months.
- `--output`: Output format: `table` (default) or `json`.
- `--waste`: Makes an analysis of possible money waste you have in your AWS Account.
  - [x] Unused EBS Volumes (not attached to any instance).
  - [x] EBS Volumes attached to stopped EC2 instances.
  - [x] Unassociated Elastic IPs.
  - [x] EC2 reserved instance that are scheduled to expire in the next 30 days or have expired in the preceding 30 days.
  - [x] EC2 instance stopped for more than 30 days.
  - [x] Load Balancers with no attached target groups.
  - [x] Unused AMIs (not associated with any running or stopped instance and created more than 90 days ago).
  - [x] Orphaned EBS Snapshots (source volume deleted and not used by any AMI).
  - [x] Stale EBS Snapshots (created more than 90 days ago, source volume exists and not used by any AMI).
  - [x] Unused EC2 Key Pairs (not associated with any running or stopped instance).
  - [x] S3 Buckets without lifecycle policies.
  - [ ] Inactive VPC interface endpoints.
  - [ ] Inactive NAT Gateways.
  - [ ] Idle Load Balancers.
  - [ ] RDS Idle DB Instances.
- `--version`: Display version information.
- `--update`: Updates the tool to the latest version.

> [!TIP]
> If your AWS profile uses `assume_role` with `mfa_serial`, **aws-doctor** will automatically prompt you to enter your MFA token code securely.

## Roadmap

- [x] Add monthly trend analysis
- [x] Add waste / wastage analysis logic
- [x] Export reports to JSON format
- [ ] Export reports to CSV and PDF formats (medical records for your cloud)
- [ ] Distribute the CLI via Fedora, Ubuntu, and macOS repositories
