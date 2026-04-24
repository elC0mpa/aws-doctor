# aws-doctor

<p align="center">
  <a href="https://awsdoctor.compacompila.com/"><img src="https://img.shields.io/badge/Documentation-Website-blue?style=for-the-badge&logo=hugo" alt="Website"></a>
</p>

<p align="center">
  <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg" alt="awesome-go"></a>
</p>

<p align="center">
  <a href="https://github.com/elC0mpa/aws-doctor/blob/main/go.mod"><img src="https://img.shields.io/github/go-mod/go-version/elC0mpa/aws-doctor" alt="Go Version"></a>
  <a href="https://pkg.go.dev/github.com/elC0mpa/aws-doctor"><img src="https://pkg.go.dev/badge/github.com/elC0mpa/aws-doctor.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/elC0mpa/aws-doctor"><img src="https://goreportcard.com/badge/github.com/elC0mpa/aws-doctor" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/elC0mpa/aws-doctor"><img src="https://codecov.io/gh/elC0mpa/aws-doctor/graph/badge.svg" alt="codecov"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/releases"><img src="https://img.shields.io/github/downloads/elC0mpa/aws-doctor/total?color=blue&label=Downloads" alt="GitHub all releases"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/actions/workflows/ci.yml"><img src="https://github.com/elC0mpa/aws-doctor/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/elC0mpa/aws-doctor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/elC0mpa/aws-doctor" alt="License"></a>
</p>

A terminal-based tool that acts as a comprehensive health check for your AWS accounts. Built with Golang, **aws-doctor** diagnoses cost anomalies, detects idle resources, and provides a proactive analysis of your cloud infrastructure.

> [!TIP]
> **View the full documentation, permissions guide, and usage examples at [awsdoctor.compacompila.com](https://awsdoctor.compacompila.com/)**

## 👀 Quick glance

### ⚖️ Comparative Cost Analytics

![Comparative Cost Analytics](https://github.com/elC0mpa/aws-doctor/blob/main/docs/static/images/demo/basic.gif?raw=true)

### 📈 6-Month Trend Analysis

![6-Month Trend Analysis](https://github.com/elC0mpa/aws-doctor/blob/main/docs/static/images/demo/trend.gif?raw=true)

### 🧟 Waste Detection

![Waste Detection](https://github.com/elC0mpa/aws-doctor/blob/main/docs/static/images/demo/waste.gif?raw=true)

_Supports selective scanning: `aws-doctor waste ec2 s3 cloudwatch rds vpc lambda sagemaker elb`_

## 📄 Professional Reporting

`aws-doctor` can now generate detailed, professional PDF reports ready for stakeholders. Reports include branded headers, styled tables, and comprehensive cost/waste analyses.

> [!TIP]
> **View PDF reporting examples and details at [awsdoctor.compacompila.com/docs/reporting/](https://awsdoctor.compacompila.com/docs/reporting/)**

### Generate a Cost Comparison Report

```bash
aws-doctor report cost
```

### Generate a Waste Analysis Report

```bash
# Full waste report
aws-doctor report waste

# Selective checks (e.g., ec2, s3, and lambda only)
aws-doctor report waste ec2 s3 lambda sagemaker
```

### Generate a Trend Report

```bash
# Full trend report (all services)
aws-doctor report trend

# Selective services (e.g., ec2 and rds only)
aws-doctor report trend ec2 rds
```

> [!TIP]
> **Subcommand Arguments:** Just like the terminal commands, `report waste` accepts specific checks (e.g., `ec2`, `s3`, `rds`, `lambda`) and `report trend` accepts specific service names.

> [!TIP]
> By default, reports are saved in your **Documents** folder. Use the `--path` flag to specify a custom directory or filename:
> `aws-doctor report cost --path ./billing-analysis.pdf`

## 🚀 Installation

**Homebrew (macOS/Linux):**

```bash
brew install elC0mpa/homebrew-tap/aws-doctor
```

**One-Line Script (macOS/Linux):**

```bash
curl -sSfL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.ps1 | iex
```

**Using Go:**

```bash
go install github.com/elC0mpa/aws-doctor@latest
```

## ✨ Key Features

- **📄 Professional PDF Reports:** Generate branded, ready-to-share PDF documents for costs, trends, and waste analysis.
- **📉 Fair Cost Comparison:** Compares identical time windows between months to spot real anomalies.
- **🧟 Zombie Discovery:** Scans for idle EIPs, stopped instances, orphaned snapshots, idle RDS instances, idle NAT Gateways, idle Load Balancers, over-provisioned Lambda memory, and idle SageMaker real-time inference endpoints. Supports selective service filtering (`ec2`, `s3`, `elb`, `cloudwatch`, `rds`, `vpc`, `lambda`, `sagemaker`).
- **📊 6-Month Trends:** High-fidelity ANSI visualization of your spending velocity.
- **📤 Multiple Output Formats:** Export results in `table`, `json`, or `csv` for easy integration with other tools or reporting.
- **🔔 Update Notifications:** Automatically checks for new versions in the background and notifies you after command output, so you never miss an update.
- **🔐 MFA Ready:** Native support for profiles requiring Multi-Factor Authentication.
- **🌍 Region-Aware Pricing:** Queries the AWS Pricing API at startup to use rates for the configured region, falling back to us-east-1 defaults if the API is unavailable. Requires `pricing:GetProducts` in the caller's IAM policy; without it, estimates silently fall back to defaults.

## 💡 Motivation

As a Cloud Architect, I often need to check AWS costs and billing information. While the AWS Console provides raw data, it lacks the immediate context I need to answer the question: _*"Are we spending efficiently?"*_

I created **\*\*aws-doctor\*\*** to fill that gap. It doesn't just show you the bill; it acts as a diagnostic tool that helps you understand **\*\*where\*\*** the money is going and **\*\*what\*\*** can be cleaned up. It automates the routine checks I used to perform manually, serving as a free, open-source alternative to the paid recommendations found in AWS Trusted Advisor.

## 👥 Community

### Contributors

A huge thank you to everyone who has contributed to **aws-doctor**! Your help makes this tool better for everyone.

<a href="https://github.com/elC0mpa/aws-doctor/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=elC0mpa/aws-doctor" alt="Contributors" />
</a>

### Star History

<a href="https://www.star-history.com/?type=date&legend=top-left&repos=elC0mpa%2Faws-doctor">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=elC0mpa/aws-doctor&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=elC0mpa/aws-doctor&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=elC0mpa/aws-doctor&type=date&legend=top-left" />
 </picture>
</a>

## 🤝 Contributing

We love contributions! Whether it's a new detection rule or a bug fix, check our [Community Dashboard](https://awsdoctor.compacompila.com/#join-the-community) to get started.

> [!IMPORTANT]
> **Always target your Pull Requests to the `development` branch.** The `main` branch is reserved for production-ready releases. Check our [Contributing Guidelines](CONTRIBUTING.md) for more details.
