---
title: "Cost Analytics"
description: "Discover the two primary cost analysis workflows: Comparative and Trend analysis."
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** focuses on contextual cost analysis. Instead of just showing raw numbers, it helps you understand how your spending is evolving.

## Comparative Workflow

To get a side-by-side comparison of your spending, use the `cost` subcommand:

```bash
aws-doctor cost
```

This triggers the **Comparative Workflow**, which includes a per-service breakdown (EC2, S3, etc.) to help you identify specific cost drivers.

![Comparative Cost Analytics](/images/demo/basic.gif)

### Fairness in Comparison
Traditional billing comparisons (like Month-over-Month) are often misleading. Comparing the 15th of October to the full month of September will always look like a "saving," even if you are spending more.

**AWS Doctor** solves this by comparing identical time windows:
- **Current Period**: 1st day of the month → Today.
- **Previous Period**: 1st day of the previous month → Identical day last month.

*Example: If today is Oct 15th, it compares Oct 1-15 vs Sept 1-15.*

{{< callout type="warning" >}}
**1st Day of the Month**: This feature is not available on the first day of the month. AWS Cost Explorer requires a minimum 24-hour range where the start date is strictly before the end date.
{{< /callout >}}

---

## 6-Month Trend Analysis

To spot long-term growth patterns or sudden architectural shifts, use the `trend` subcommand:

```bash
aws-doctor trend
```

![6-Month Trend Analysis](/images/demo/trend.gif)

### Filtering by Service
You can now focus your trend analysis on specific AWS services by passing them as arguments to the `trend` subcommand. This is useful for monitoring the cost evolution of a single service or a group of services.

```bash
# Analyze cost trend for EC2 and S3 only
aws-doctor trend ec2 s3
```

You can pass multiple services separated by spaces or commas.

### Available Service Shorthands
**AWS Doctor** uses intuitive shorthands for filtering. Each shorthand maps to its official AWS Cost Explorer service name:

| Shorthand | AWS Service | Shorthand | AWS Service |
| :--- | :--- | :--- | :--- |
| `ec2` | EC2 Compute | `s3` | S3 Storage |
| `rds` | RDS Database | `lambda` | AWS Lambda |
| `dynamodb` | DynamoDB | `eks` | EKS (Kubernetes) |
| `ecs` | ECS (Containers) | `elb` | Elastic Load Balancing |
| `vpc` | Virtual Private Cloud | `route53` | Route 53 |
| `apigateway` | API Gateway | `cloudfront` | CloudFront |
| `cloudwatch` | CloudWatch | `elasticache` | ElastiCache |
| `redshift` | Redshift | `savingsplans` | Savings Plans |
| `glue` | AWS Glue | `kinesis` | Kinesis |
| `firehose` | Kinesis Firehose | `quicksight` | QuickSight |
| `waf` | AWS WAF | `backup` | AWS Backup |
| `stepfunctions`| Step Functions | `kms` | KMS |
| `secretsmanager`| Secrets Manager | `ssm` | Systems Manager |
| `location` | Location Service | `ecr` | ECR (Registry) |

### What it shows:
- A high-fidelity ANSI bar chart in your terminal.
- Monthly total costs for the last 6 completed billing cycles.
- Clear indicators of whether your spend is accelerating or stabilizing.
