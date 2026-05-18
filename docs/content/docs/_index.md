---
title: "Documentation"
description: "Comprehensive guide to using AWS Doctor. Learn how to install, configure permissions, and use the tool to audit your AWS infrastructure for costs and security."
weight: 1
---

Welcome to the **AWS Doctor** documentation. This guide will help you set up, configure, and master the tool to keep your AWS infrastructure lean, secure, and cost-effective.

## Navigation

<div class="hx:mb-6"></div>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="terminal"
    title="Basics & Setup"
    link="getting-started/"
    subtitle="Learn how to install AWS Doctor and configure the minimum required permissions."
  >}}
  {{< hextra/feature-card
    icon="key"
    title="Usage Guide"
    link="usage/"
    subtitle="Detailed explanation of CLI subcommands, global flags, MFA support, and profile management."
  >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Cost Analytics"
    link="cost-analytics/"
    subtitle="Understand how AWS Doctor performs fair cost comparisons and trend analysis."
  >}}
  {{< hextra/feature-card
    icon="search"
    title="Waste"
    link="waste-detection/"
    subtitle="In-depth technical logic for detecting waste in EC2, S3, and Networking."
  >}}
  {{< hextra/feature-card
    icon="server"
    title="Automation"
    link="automation/"
    subtitle="Guide to JSON output and integration with GitHub Actions or Jenkins."
  >}}
  {{< hextra/feature-card
    icon="printer"
    title="PDF Reporting"
    link="reporting/"
    subtitle="Generate professional, brandable PDF reports for stakeholders with one command."
  >}}
{{< /hextra/feature-grid >}}

## Quick Context

- **Fast & Efficient**: Implements local caching to minimize network calls.
- **Stateless & Secure**: Never stores your AWS credentials or sensitive account data.
- **Fair Assessment**: Cost comparisons use identical time windows for accuracy.
- **Zero Config**: Works out-of-the-box with your existing `~/.aws/config`.

{{< callout type="info" >}}
Looking for something specific? Use the search bar at the top of the page to find details about a particular service or subcommand.
{{< /callout >}}
