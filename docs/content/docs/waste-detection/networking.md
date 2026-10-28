---
title: "Networking"
description: "Detect unassociated Elastic IPs and idle Load Balancers without healthy targets to eliminate unnecessary networking costs."
weight: 30
---

Uncover costs from unattached networking assets and idle connectivity resources.

{{< callout type="info" >}}
**Permissions Required**: `ec2:DescribeAddresses`, `ec2:DescribeNetworkInterfaces`, `ec2:DescribeNatGateways`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Elastic IP Addresses (EIP)

**AWS Doctor** identifies EIPs that are not currently associated with an instance or network interface.

### The Cost of Idle IPs
AWS charges for all public IPv4 addresses, including Elastic IPs. While an associated IP provides connectivity, an **unassociated** (idle) EIP is pure waste—you are paying the hourly rate for a resource that isn't providing any value to your infrastructure.

- **Action**: Release any EIP that isn't actively mapped to a service.

---

## Public IPv4 Charges (in use)

Since February 2024, AWS charges the hourly public IPv4 rate (~$0.005/hour, ~$3.65/month) for **every** public IPv4 address, not only idle Elastic IPs. **AWS Doctor** lists the public IPv4 addresses currently attached to network interfaces and summarizes them in a **Public IPv4 Charges (in use)** table with one row per resource type (EC2 Instance, NAT Gateway, Load Balancer, RDS Database, Other), the number of addresses, and the estimated monthly cost.

- **Informational only**: Most of these addresses are legitimately needed, so they are **not** counted as waste, are not part of the **Total Estimated Monthly Waste**, and do not change the `has_waste` value in JSON output. In JSON the summary is the `in_use_public_ipv4_charges` array.
- **Action**: Use the figure to decide where IPv6, VPC endpoints, or fewer public addresses could reduce your bill.
- **Estimate**: The figure is `address count x hourly rate x 730 hours`. Bring-your-own-IP (BYOIP) addresses are not billed by AWS but are still counted, and the 750-hour monthly free tier for new accounts is not subtracted.
- **Scope**: Elastic IPs that are not associated with anything do not appear here; they are reported in the **Elastic IP Waste** table above.

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

{{< callout type="tip" >}}
You can customize the 7-day idle threshold using the `--elb-idle-days` flag.
{{< /callout >}}


---

## NAT Gateways

**AWS Doctor** identifies NAT Gateways that have processed **zero bytes** of data over the last **7 days**.

### Why it's waste
NAT Gateways have a high hourly cost (~$32.85/month in most regions) even when they are not processing any traffic. If a NAT Gateway is idle, it is often a leftover from a previous architecture or an improperly decommissioned environment.

- **Action**: Delete any NAT Gateway that shows no activity and is no longer required for outbound connectivity.

{{< callout type="tip" >}}
You can customize the 7-day idle threshold using the `--vpc-nat-idle-days` flag.
{{< /callout >}}

{{< callout type="info" >}}
Future updates will include detection for **Unused VPC Endpoints**.
{{< /callout >}}
