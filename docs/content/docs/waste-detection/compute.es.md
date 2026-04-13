---
title: "Computo y EBS"
description: "Audite instancias EC2, volumenes EBS, snapshots y funciones Lambda en busca de desperdicio. Identifique instancias detenidas, almacenamiento huerfano y funciones sobreaprovisionadas para ahorrar costos."
weight: 10
---

Audite su huella de EC2, EBS y Lambda para eliminar los costos de instancias, datos abandonados y funciones sobreaprovisionadas.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeInstances`, `ec2:DescribeReservedInstances`, `ec2:DescribeVolumes`, `ec2:DescribeSnapshots`, `ec2:DescribeKeyPairs`, `ec2:DescribeImages`, `lambda:ListFunctions`, `logs:FilterLogEvents`.
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
Identifica los Key Pairs de EC2 que no estan asociados con ninguna instancia en ejecucion o detenida.
- **Razon**: Reduce el desorden administrativo y los posibles riesgos de seguridad de llaves antiguas.
- **Accion**: Eliminar las llaves sin usar desde la consola/CLI.

---

## Funciones Lambda

### Memoria Sobreaprovisionada
Identifica funciones Lambda que consistentemente usan **menos del 10%** de su memoria asignada durante los ultimos 14 dias. La deteccion funciona analizando `Max Memory Used` de las lineas `REPORT` en los CloudWatch Logs de cada funcion.

- **Razon**: El precio de Lambda se basa en GB-segundos. Una funcion con 1024 MB asignados pero que solo usa 50 MB esta pagando aproximadamente 20x mas por invocacion de lo necesario. Ajustar la memoria puede reducir significativamente los costos.
- **Accion**: Reducir la asignacion de memoria de la funcion. AWS Doctor sugiere un tamano recomendado de 2x el uso pico observado (minimo 128 MB) para permitir margen para picos.

El umbral predeterminado del 10% se puede personalizar:

```bash
aws-doctor waste lambda --lambda-memory-threshold 20
```
