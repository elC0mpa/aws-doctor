---
title: "AWS Doctor"
description: "AWS Doctor is a powerful open-source CLI tool to audit security, costs, and best practices in AWS. Identify cloud waste and optimize your infrastructure easily."
layout: "hextra-home"
---

{{< hextra/hero-container
  image="/images/logo.webp"
  imageTitle="AWS Doctor"
  imageWidth="512"
>}}
{{< hextra/hero-badge link="https://github.com/elC0mpa/aws-doctor/releases" >}}
  <div class="hx-w-2 hx-h-2 hx-rounded-full hx-bg-primary-400"></div>
  <span>Latest version: {{< latest-version >}}</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6 hx:mt-6">
{{< hextra/hero-headline >}}
  AWS Doctor
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mt-6 hx-mb-6">
{{< hextra/hero-subtitle >}}
  Powerful open-source CLI to audit security, costs, and best practices in AWS.
{{< /hextra/hero-subtitle >}}
</div>

{{< hero-buttons >}}
{{< hextra/hero-button text="Get Started" link="docs/" >}}
{{< hextra/hero-badge style="display: flex; justify-content: center; padding: 13px 12px !important; font-size: .875rem !important;" link="https://github.com/elC0mpa/aws-doctor" >}}
  <span>View on GitHub <img class="not-prose" style="display: inline; height: 22px; margin-left: 8px;" src='https://img.shields.io/github/stars/elC0mpa/aws-doctor?style=social'/></span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
{{< /hero-buttons >}}
{{< /hextra/hero-container >}}

<div class="hx:mt-12"></div>

{{< hextra/hero-section >}}
  Core Features
{{< /hextra/hero-section >}}

<div class="hx:mt-4"></div>

{{< hextra/feature-grid cols="4" >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Cost Analytics"
    subtitle="Gain a fair assessment of your spending velocity. AWS Doctor compares your current month's costs against the exact same period in the previous month (e.g., 1st–10th), allowing you to spot anomalies and spikes in real-time."
  >}}

  {{< hextra/feature-card
    icon="trash"
    title="Zombie Discovery"
    subtitle="Get a high-level health check of your entire AWS account. The tool scans multiple services simultaneously to identify idle, unattached, and forgotten resources, providing a unified view of infrastructure waste in seconds."
  >}}

  {{< hextra/feature-card
    icon="terminal"
    title="Output Formats"
    subtitle="Choose the format that fits your workflow. Use rich terminal tables for quick manual audits, or generate structured JSON output to feed data into your CI/CD pipelines, custom dashboards, and automation scripts."
  >}}

  {{< hextra/feature-card
    icon="key"
    title="Security & IAM"
    subtitle="Full support for MFA-protected roles and proactive IAM credential audits."
  >}}

{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Instant Infrastructure Audit
{{< /hextra/hero-section >}}

<div class="hx:mt-4"></div>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="server"
    title="Compute & EBS"
    subtitle="Detect idle EC2 instances, unattached EBS volumes, and orphaned snapshots."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="S3 Storage"
    subtitle="Audit buckets without lifecycle policies and cleanup abandoned multipart uploads."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Networking"
    subtitle="Identify unassociated Elastic IPs and Load Balancers without healthy targets."
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Join the Community
{{< /hextra/hero-section >}}

{{< repo-stats >}}

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="terminal"
    title="Report Issues"
    subtitle="Found a bug or have an idea for a new detection rule? Help us improve the tool by opening an issue on GitHub."
    link="https://github.com/elC0mpa/aws-doctor/issues"
  >}}
  {{< hextra/feature-card
    icon="github"
    title="Contribute Code"
    subtitle="Ready to contribute? We welcome PRs for new features, bug fixes, and documentation improvements."
    link="https://github.com/elC0mpa/aws-doctor/pulls"
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-24"></div>
