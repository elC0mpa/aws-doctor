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

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Cost Analytics"
    subtitle="Gain a fair assessment of your spending velocity. AWS Doctor compares your current month's costs against the exact same period in the previous month (e.g., 1st–10th), allowing you to spot anomalies and spikes in real-time."
    link="docs/cost-analytics/"
  >}}

  {{< hextra/feature-card
    icon="trash"
    title="Zombie Discovery"
    subtitle="Get a high-level health check of your entire AWS account. The tool scans multiple services simultaneously to identify idle, unattached, and forgotten resources, providing a unified view of infrastructure waste in seconds."
    link="docs/waste-detection/"
  >}}

  {{< hextra/feature-card
    icon="printer"
    title="PDF Reporting"
    subtitle="Generate professional, brandable PDF reports for stakeholders. AWS Doctor can now export all audit findings, cost trends, and waste summaries into a clean, ready-to-share document."
    link="docs/reporting/"
  >}}

  {{< hextra/feature-card
    icon="globe-alt"
    title="Region-Aware Pricing"
    subtitle="Cost estimates are backed by live data from the AWS Pricing API for your configured region. If the API is unavailable, the tool falls back to built-in defaults so your scan never fails."
    link="docs/waste-detection/#region-aware-cost-estimation"
  >}}

  {{< hextra/feature-card
    icon="terminal"
    title="Output Formats"
    subtitle="Choose the format that fits your workflow. Experience a rich, interactive terminal UI for manual audits, or generate structured JSON output to feed data into your CI/CD pipelines and automation scripts."
    link="docs/usage/"
  >}}

  {{< hextra/feature-card
    icon="key"
    title="Security & IAM"
    subtitle="Full support for MFA-protected roles and proactive IAM credential audits."
    link="docs/usage/#mfa-support"
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
    title="Compute and EBS"
    subtitle="Detect stopped EC2 instances, unattached EBS volumes, orphaned snapshots, unused AMIs, unused key pairs, expiring Reserved Instances, and over-provisioned Lambda memory."
  >}}
  {{< hextra/feature-card
    icon="database"
    title="Databases"
    subtitle="Identify stopped RDS instances, old manual snapshots, and idle database connections."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="Storage and Logs"
    subtitle="Audit S3 buckets without lifecycle policies, abandoned multipart uploads, CloudWatch Log Groups without retention, and ECR repositories with untagged images or missing lifecycle policies."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Networking"
    subtitle="Identify unassociated Elastic IPs, idle NAT Gateways, and Load Balancers without healthy targets."
  >}}
  {{< hextra/feature-card
    icon="chip"
    title="Machine Learning"
    subtitle="Detect idle SageMaker endpoints with zero recent invocations."
  >}}
  {{< hextra/feature-card
    icon="key"
    title="Configuration and Secrets"
    subtitle="Flag unused Secrets Manager secrets that have not been accessed within a configurable threshold."
  >}}
  {{< hextra/feature-card
    icon="shield"
    title="Identity & Security"
    subtitle="Unused IAM users and Root accounts without MFA."
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
