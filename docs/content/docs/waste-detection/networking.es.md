---
title: "Redes"
description: "Detecte IPs Elásticas sin asociar y Load Balancers inactivos sin objetivos saludables para eliminar costos de red innecesarios."
weight: 30
---

Descubra los costos de los activos de red desconectados y los recursos de conectividad inactivos.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeAddresses`, `ec2:DescribeNatGateways`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Direcciones IP Elásticas (EIP)

**AWS Doctor** identifica las EIP que no están asociadas actualmente con una instancia o interfaz de red.

### El Costo de las IPs Inactivas
AWS cobra por todas las direcciones IPv4 públicas, incluyendo las IPs Elásticas. Mientras que una IP asociada proporciona conectividad, una EIP **sin asociar** (inactiva) es puro desperdicio: está pagando la tarifa por hora por un recurso que no proporciona ningún valor a su infraestructura.

- **Acción**: Liberar cualquier EIP que no esté mapeada activamente a un servicio.

---

## Elastic Load Balancers (ELB)

Identifica los Application (ALB) y Network (NLB) Load Balancers que **no están asociados con ningún grupo de destino (target group)**.

### Por qué es desperdicio
Los Load Balancers tienen un costo fijo por hora independientemente del volumen de tráfico. Un ELB sin grupos de destino es efectivamente un punto de entrada a ninguna parte, pero sigue facturando a la tarifa por hora completa más los cargos por LCU.

- **Acción**: Eliminar cualquier Load Balancer que tenga cero objetivos saludables o no tenga una asociación de grupo de destino.

---

## NAT Gateways

**AWS Doctor** identifica los NAT Gateways que han procesado **cero bytes** de datos durante los últimos **7 días**.

### Por qué es desperdicio
Los NAT Gateways tienen un alto costo por hora (~$32.85/mes en la mayoría de las regiones) incluso cuando no están procesando ningún tráfico. Si un NAT Gateway está inactivo, a menudo es un resto de una arquitectura anterior o de un entorno desmantelado incorrectamente.

- **Acción**: Eliminar cualquier NAT Gateway que no muestre actividad y que ya no sea necesario para la conectividad de salida.

{{< callout type="info" >}}
Las futuras actualizaciones incluirán la detección de **VPC Endpoints sin usar**.
{{< /callout >}}
