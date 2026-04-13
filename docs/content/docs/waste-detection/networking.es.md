---
title: "Redes"
description: "Detecte IPs Elásticas sin asociar, Load Balancers inactivos y NAT Gateways inactivos para eliminar costos de red innecesarios."
weight: 30
---

Descubra los costos de los activos de red desconectados y los recursos de conectividad inactivos.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeAddresses`, `ec2:DescribeNatGateways`, `cloudwatch:GetMetricStatistics`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`.
{{< /callout >}}

## Direcciones IP Elásticas (EIP)

**AWS Doctor** identifica las EIP que no están asociadas actualmente con una instancia o interfaz de red.

### El Costo de las IPs Inactivas
AWS cobra por todas las direcciones IPv4 públicas, incluyendo las IPs Elásticas. Mientras que una IP asociada proporciona conectividad, una EIP **sin asociar** (inactiva) es puro desperdicio - esta pagando la tarifa por hora por un recurso que no proporciona ningun valor a su infraestructura.

- **Accion**: Liberar cualquier EIP que no este mapeada activamente a un servicio.

---

## Elastic Load Balancers (ELB)

Identifica los Application (ALB) y Network (NLB) Load Balancers que **no estan asociados con ningun grupo de destino (target group)**.

### Por que es desperdicio
Los Load Balancers tienen un costo fijo por hora independientemente del volumen de trafico. Un ELB sin grupos de destino es efectivamente un punto de entrada a ninguna parte, pero sigue facturando a la tarifa por hora completa mas los cargos por LCU.

- **Accion**: Eliminar cualquier Load Balancer que tenga cero objetivos saludables o no tenga una asociacion de grupo de destino.

---

## NAT Gateways

Identifica NAT Gateways en estado `available` que han procesado **cero bytes de trafico saliente** durante los ultimos 7 dias, utilizando la metrica `BytesOutToDestination` de CloudWatch.

### Por que es desperdicio
Los NAT Gateways tienen un costo fijo por hora de aproximadamente **$0.045/hora (~$32.85/mes)** en us-east-1, mas cargos por procesamiento de datos. Un NAT Gateway inactivo sin trafico fluyendo a traves de el no proporciona ningun valor mientras sigue incurriendo en el cargo por hora completo.

- **Accion**: Eliminar cualquier NAT Gateway que ya no este sirviendo trafico. Verifique que ningun recurso en la subred privada asociada requiera acceso a internet saliente antes de eliminar.

{{< callout type="info" >}}
Las futuras actualizaciones incluiran la deteccion de **VPC Endpoints sin usar**.
{{< /callout >}}
