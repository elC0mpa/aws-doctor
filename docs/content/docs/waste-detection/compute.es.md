---
title: "Cómputo y EBS"
description: "Audite instancias EC2, volúmenes EBS y snapshots en busca de desperdicio. Identifique instancias detenidas y almacenamiento huérfano para ahorrar costos."
weight: 10
---

Audite su huella de EC2 y EBS para eliminar los costos de instancias y datos abandonados.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeInstances`, `ec2:DescribeReservedInstances`, `ec2:DescribeVolumes`, `ec2:DescribeSnapshots`, `ec2:DescribeKeyPairs`, `ec2:DescribeImages`.
{{< /callout >}}

## Instancias EC2

### Instancias Detenidas por Mucho Tiempo
**AWS Doctor** identifica las instancias que han estado en estado `stopped` durante **más de 30 días**.
- **Razón**: Aunque no paga por CPU/RAM cuando están detenidas, sigue pagando por los volúmenes raíz de EBS asociados y cualquier almacenamiento persistente.
- **Acción**: Terminar o realizar un snapshot de los datos y eliminar.

### Reserved Instances (RI) por Vencer
Escanea RIs activas programadas para vencer en los **próximos 30 días** o que han vencido en los **últimos 30 días**.
- **Razón**: Las RIs vencidas vuelven a los costosos precios de On-Demand sin previo aviso.
- **Acción**: Revisar el uso y renovar o migrar a Savings Plans.

---

## Volúmenes y Snapshots de EBS

### Volúmenes EBS sin Usar
Encuentra volúmenes con un estado de `available` (lo que significa que no están conectados a ninguna instancia).
- **Razón**: Se le factura por el tamaño aprovisionado de estos volúmenes cada hora que existen.
- **Acción**: Eliminar si ya no son necesarios.

### Snapshots Huérfanos
Encuentra snapshots donde el **volumen de origen ha sido eliminado** y el snapshot no está asociado con ninguna AMI.
- **Razón**: A menudo creados durante copias de seguridad manuales o despliegues antiguos y olvidados.
- **Acción**: Eliminar para ahorrar en costos de almacenamiento respaldados por S3.

### Snapshots y AMIs Obsoletos
Marca las AMIs y snapshots que tienen **más de 90 días** y no están asociados con ninguna instancia en ejecución o detenida.
- **Razón**: Imágenes base y copias de seguridad desactualizadas que probablemente no se han tocado en un trimestre.
- **Acción**: Limpiar versiones antiguas de imágenes.

---

## Acceso y Seguridad

### Key Pairs sin Usar
Identifica los Key Pairs de EC2 que no están asociados con ninguna instancia en ejecución o detenida.
- **Razón**: Reduce el desorden administrativo y los posibles riesgos de seguridad de llaves antiguas.
- **Acción**: Eliminar las llaves sin usar desde la consola/CLI.
