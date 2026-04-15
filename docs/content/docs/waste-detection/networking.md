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

Identifies Application (ALB) and Network (NLB) Load Balancers that are **not associated with any target group**.

### Why it's waste
Load Balancers carry a fixed hourly cost regardless of traffic volume. An ELB without target groups is effectively an entry point to nowhere, yet it continues to bill at the full hourly rate plus LCU charges.

- **Action**: Delete any Load Balancer that has zero healthy targets or no target group association.

---

## NAT Gateways

**AWS Doctor** identifies NAT Gateways that have processed **zero bytes** of data over the last **7 days**.

### Why it's waste
NAT Gateways have a high hourly cost (~$32.85/month in most regions) even when they are not processing any traffic. If a NAT Gateway is idle, it is often a leftover from a previous architecture or an improperly decommissioned environment.

- **Action**: Delete any NAT Gateway that shows no activity and is no longer required for outbound connectivity.

{{< callout type="info" >}}
Future updates will include detection for **Unused VPC Endpoints**.
{{< /callout >}}
