---
title: "Networking"
description: "Detect unassociated Elastic IPs, idle Load Balancers, and idle NAT Gateways to eliminate unnecessary networking costs."
weight: 30
---

Uncover costs from unattached networking assets and idle connectivity resources.

{{< callout type="info" >}}
**Permissions Required**: `ec2:DescribeAddresses`, `ec2:DescribeNatGateways`, `cloudwatch:GetMetricStatistics`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`.
{{< /callout >}}

## Elastic IP Addresses (EIP)

**AWS Doctor** identifies EIPs that are not currently associated with an instance or network interface.

### The Cost of Idle IPs
AWS charges for all public IPv4 addresses, including Elastic IPs. While an associated IP provides connectivity, an **unassociated** (idle) EIP is pure waste - you are paying the hourly rate for a resource that isn't providing any value to your infrastructure.

- **Action**: Release any EIP that isn't actively mapped to a service.

---

## Elastic Load Balancers (ELB)

Identifies Application (ALB) and Network (NLB) Load Balancers that are **not associated with any target group**.

### Why it's waste
Load Balancers carry a fixed hourly cost regardless of traffic volume. An ELB without target groups is effectively an entry point to nowhere, yet it continues to bill at the full hourly rate plus LCU charges.

- **Action**: Delete any Load Balancer that has zero healthy targets or no target group association.

---

## NAT Gateways

Identifies NAT Gateways in the `available` state that have processed **zero bytes of outbound traffic** over the last 7 days, using the `BytesOutToDestination` CloudWatch metric.

### Why it's waste
NAT Gateways carry a fixed hourly cost of approximately **$0.045/hour (~$32.85/month)** in us-east-1, plus data processing charges. An idle NAT Gateway with no traffic flowing through it provides no value while continuing to incur the full hourly charge.

- **Action**: Remove any NAT Gateway that is no longer serving traffic. Verify that no resources in the associated private subnet require outbound internet access before deleting.

{{< callout type="info" >}}
Future updates will include detection for **Unused VPC Endpoints**.
{{< /callout >}}
