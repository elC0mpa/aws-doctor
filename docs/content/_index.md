---
title: "AWS Doctor"
layout: "hextra-home"
---

{{< hextra/hero-container
  image="/images/logo.webp"
  imageTitle="AWS Doctor"
  imageWidth="512"
>}}
{{< hextra/hero-badge link="https://github.com/elC0mpa/aws-doctor/releases" >}}
  <div class="hx-w-2 hx-h-2 hx-rounded-full hx-bg-primary-400"></div>
  <span>Latest version: v1.6.0</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6 hx:mt-4">
{{< hextra/hero-headline >}}
  AWS Doctor
{{< /hextra/hero-headline >}}
</div>

<div class="hx-mt-4 hx-mb-6">
{{< hextra/hero-subtitle >}}
  A powerful, open-source CLI tool to identify security risks, cost optimizations, and operational best practices in your cloud environment.
{{< /hextra/hero-subtitle >}}
</div>

<div class="hx:mt-4 hx:flex hx:gap-x-2">
{{< hextra/hero-button text="Get Started" link="docs/" >}}
{{< hextra/hero-badge style="padding: 13px 12px !important; font-size: .875rem !important;" link="https://github.com/elC0mpa/aws-doctor" >}}
  <span>View on GitHub <img class="not-prose" style="display: inline; height: 22px; margin-left: 8px;" src='https://img.shields.io/github/stars/elC0mpa/aws-doctor?style=social'/></span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
</div>
{{< /hextra/hero-container >}}

<div class="hx:mt-16"></div>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Comparative Cost Analytics"
    subtitle="Gain a fair assessment of your spending velocity. AWS Doctor compares your current month's costs against the exact same period in the previous month (e.g., 1st–10th), allowing you to spot anomalies and spikes in real-time."
  >}}

  {{< hextra/feature-card
    icon="trash"
    title="Zombie Resource Discovery"
    subtitle="Get a high-level health check of your entire AWS account. The tool scans multiple services simultaneously to identify idle, unattached, and forgotten resources, providing a unified view of infrastructure waste in seconds."
  >}}

  {{< hextra/feature-card
    icon="terminal"
    title="Flexible Output Formats"
    subtitle="Choose the format that fits your workflow. Use rich terminal tables for quick manual audits, or generate structured JSON output to feed data into your CI/CD pipelines, custom dashboards, and automation scripts."
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Virtual Tour
{{< /hextra/hero-section >}}

{{< columns cols="2" >}}
  {{< column
      title="Ongoing Cost Comparison"
      border="true"
      image="/images/demo/basic.gif"
  >}}
    Fair assessment of spending velocity by comparing the current month against the previous period. Spot spikes and anomalies instantly.
  {{< /column >}}

  {{< column
      title="6-Month Trend Analysis"
      border="true"
      image="/images/demo/trend.gif"
  >}}
    Visualize cost history to spot long-term anomalies and infrastructure trends. Understand your growth patterns over time.
  {{< /column >}}

  {{< column
      title="Intelligent Waste Detection"
      border="true"
      image="/images/demo/waste.gif"
  >}}
    Scans your account for 'zombie' resources like unattached volumes, idle instances, and abandoned S3 multipart uploads.
  {{< /column >}}

  {{< column
      title="Multi-Service Support"
      border="true"
      image="/images/logo.webp"
      imageStyle="width: 128px; margin: 2rem auto;"
  >}}
    Comprehensive support for EC2, EBS, S3, ELB, and more. One tool to provide a bird's-eye view of your entire AWS infrastructure.
  {{< /column >}}
{{< /columns >}}

<div class="hx-mt-20"></div>
