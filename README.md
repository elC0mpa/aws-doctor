# aws-doctor

<p align="center">
  <a href="https://awsdoctor.compacompila.com/"><img src="https://img.shields.io/badge/Documentation-Website-blue?style=for-the-badge&logo=hugo" alt="Website"></a>
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

## 🏥 Quick Scan

![](https://github.com/elC0mpa/aws-doctor/blob/main/docs/static/images/demo/waste.gif)

## 🚀 Installation

**One-Line Script (macOS/Linux):**

```bash
curl -sSfL https://raw.githubusercontent.com/elC0mpa/aws-doctor/main/install.sh | sh
```

**Using Go:**

```bash
go install github.com/elC0mpa/aws-doctor@latest
```

## ✨ Key Features

- **📉 Fair Cost Comparison:** Compares identical time windows between months to spot real anomalies.
- **🧟 Zombie Discovery:** Scans for idle EIPs, stopped instances, orphaned snapshots, and more.
- **📊 6-Month Trends:** High-fidelity ANSI visualization of your spending velocity.
- **🔐 MFA Ready:** Native support for profiles requiring Multi-Factor Authentication.

## 💡 Motivation

As a Cloud Architect, I often need to check AWS costs and billing information. While the AWS Console provides raw data, it lacks the immediate context I need to answer the question: *_"Are we spending efficiently?"_*

I created ***\*aws-doctor\**** to fill that gap. It doesn't just show you the bill; it acts as a diagnostic tool that helps you understand ***\*where\**** the money is going and ***\*what\**** can be cleaned up. It automates the routine checks I used to perform manually, serving as a free, open-source alternative to the paid recommendations found in AWS Trusted Advisor.

## 🤝 Contributing

We love contributions! Whether it's a new detection rule or a bug fix, check our [Community Dashboard](https://awsdoctor.compacompila.com/#join-the-community) to get started.
