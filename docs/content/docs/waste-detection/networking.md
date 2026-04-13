---
title: "Networking"
description: "Detect unassociated Elastic IPs, orphaned Load Balancers, and idle Load Balancers with zero traffic to eliminate unnecessary networking costs."
weight: 30
---

Uncover costs from unattached networking assets and idle connectivity resources.

{{< callout type="info" >}}
**Permissions Required**: `ec2:DescribeAddresses`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Elastic IP Addresses (EIP)

**AWS Doctor** identifies EIPs that are not currently associated with an instance or network interface.

### The Cost of Idle IPs
AWS charges for all public IPv4 addresses, including Elastic IPs. While an associated IP provides connectivity, an **unassociated** (idle) EIP is pure waste - you are paying the hourly rate for a resource that isn't providing any value to your infrastructure.

- **Action**: Release any EIP that isn't actively mapped to a service.

---

## Elastic Load Balancers (ELB)

### No Target Groups

Identifies Application (ALB) and Network (NLB) Load Balancers that are **not associated with any target group**. An ELB without target groups is effectively an entry point to nowhere, yet it continues to bill at the full hourly rate.

- **Action**: Delete any Load Balancer that has no target group association.

### Idle Load Balancers (Zero Traffic)

Identifies ALBs and NLBs that **have target groups but receive zero traffic** over the last 7 days. This check uses CloudWatch metrics to detect inactivity:

- **ALBs**: `RequestCount` metric in the `AWS/ApplicationELB` namespace
- **NLBs**: `ActiveFlowCount` metric in the `AWS/NetworkELB` namespace

### Why it's waste
Load Balancers carry a fixed hourly cost regardless of traffic volume. An ALB costs approximately **$16.43/month** and an NLB similarly, even with zero requests flowing through them. A load balancer with target groups but no traffic may indicate a decommissioned service that was never fully cleaned up.

- **Action**: Verify the load balancer is no longer needed, then delete it along with its target groups.

{{< callout type="info" >}}
Future updates will include detection for **Idle NAT Gateways** and **Unused VPC Endpoints**.
{{< /callout >}}
