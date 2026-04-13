---
title: "Redes"
description: "Detecte IPs Elasticas sin asociar, Load Balancers huerfanos y Load Balancers inactivos sin trafico para eliminar costos de red innecesarios."
weight: 30
---

Descubra los costos de los activos de red desconectados y los recursos de conectividad inactivos.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeAddresses`, `elasticloadbalancing:DescribeLoadBalancers`, `elasticloadbalancing:DescribeTargetGroups`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Direcciones IP Elasticas (EIP)

**AWS Doctor** identifica las EIP que no estan asociadas actualmente con una instancia o interfaz de red.

### El Costo de las IPs Inactivas
AWS cobra por todas las direcciones IPv4 publicas, incluyendo las IPs Elasticas. Mientras que una IP asociada proporciona conectividad, una EIP **sin asociar** (inactiva) es puro desperdicio - esta pagando la tarifa por hora por un recurso que no proporciona ningun valor a su infraestructura.

- **Accion**: Liberar cualquier EIP que no este mapeada activamente a un servicio.

---

## Elastic Load Balancers (ELB)

### Sin Grupos de Destino

Identifica los Application (ALB) y Network (NLB) Load Balancers que **no estan asociados con ningun grupo de destino (target group)**. Un ELB sin grupos de destino es efectivamente un punto de entrada a ninguna parte, pero sigue facturando a la tarifa por hora completa.

- **Accion**: Eliminar cualquier Load Balancer que no tenga una asociacion de grupo de destino.

### Load Balancers Inactivos (Cero Trafico)

Identifica ALBs y NLBs que **tienen grupos de destino pero reciben cero trafico** durante los ultimos 7 dias. Esta verificacion utiliza metricas de CloudWatch para detectar inactividad:

- **ALBs**: metrica `RequestCount` en el namespace `AWS/ApplicationELB`
- **NLBs**: metrica `ActiveFlowCount` en el namespace `AWS/NetworkELB`

### Por que es desperdicio
Los Load Balancers tienen un costo fijo por hora independientemente del volumen de trafico. Un ALB cuesta aproximadamente **$16.43/mes** y un NLB de manera similar, incluso con cero solicitudes fluyendo a traves de ellos. Un load balancer con grupos de destino pero sin trafico puede indicar un servicio decomisionado que nunca fue completamente limpiado.

- **Accion**: Verificar que el load balancer ya no es necesario, luego eliminarlo junto con sus grupos de destino.

{{< callout type="info" >}}
Las futuras actualizaciones incluiran la deteccion de **NAT Gateways inactivos** y **VPC Endpoints sin usar**.
{{< /callout >}}
