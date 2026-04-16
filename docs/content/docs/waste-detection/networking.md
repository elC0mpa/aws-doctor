---
title: "Networking"
description: "Detect unassociated Elastic IPs and idle Load Balancers without healthy targets to eliminate unnecessary networking costs."
weight: 30
---

Uncover costs from unattached networking assets and idle connectivity resources.

{{< callout type="info" >}}
**Permissions Required**: `ec2:DescribeAddresses`, `ec2:DescribeNatGateways`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Elastic IP Addresses (EIP)

**AWS Doctor** identifies EIPs that are not currently associated with an instance or network interface.

### The Cost of Idle IPs
AWS charges for all public IPv4 addresses, including Elastic IPs. While an associated IP provides connectivity, an **unassociated** (idle) EIP is pure waste—you are paying the hourly rate for a resource that isn't providing any value to your infrastructure.

- **Action**: Release any EIP that isn't actively mapped to a service.

---

## Elastic Load Balancers (ELB)

**AWS Doctor** identifies Application (ALB) and Network (NLB) Load Balancers that are either unassociated or idle.

### Unused Load Balancers
Flags ELBs that are **not associated with any target group**.
- **Reason**: An ELB without target groups is an entry point to nowhere, yet it continues to bill at the full hourly rate.
- **Action**: Delete any Load Balancer that has no target group association.

### Idle Load Balancers
Flags ELBs that have processed **zero requests or connections** over the last **7 days**.
- **Reason**: Load Balancers carry a fixed hourly cost regardless of traffic volume (~$16-20/month base cost).
- **Action**: Delete or consolidate services into shared LBs.

---

## NAT Gateways

**AWS Doctor** identifies NAT Gateways that have processed **zero bytes** of data over the last **7 days**.

### Why it's waste
NAT Gateways have a high hourly cost (~$32.85/month in most regions) even when they are not processing any traffic. If a NAT Gateway is idle, it is often a leftover from a previous architecture or an improperly decommissioned environment.

- **Action**: Delete any NAT Gateway that shows no activity and is no longer required for outbound connectivity.

{{< callout type="info" >}}
Future updates will include detection for **Unused VPC Endpoints**.
{{< /callout >}}
