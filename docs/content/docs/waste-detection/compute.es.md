---
title: "Cómputo y EBS"
description: "Audite instancias EC2, volúmenes EBS y snapshots en busca de desperdicio. Identifique instancias detenidas y almacenamiento huérfano para ahorrar costos."
weight: 10
---

Audite su huella de EC2 y EBS para eliminar los costos de instancias y datos abandonados.

{{< callout type="info" >}}
**Permisos Requeridos**: `ec2:DescribeInstances`, `ec2:DescribeReservedInstances`, `ec2:DescribeVolumes`, `ec2:DescribeSnapshots`, `ec2:DescribeKeyPairs`, `ec2:DescribeImages`, `cloudwatch:GetMetricStatistics`, `lambda:ListFunctions`, `logs:DescribeLogGroups`, `logs:StartQuery`, `logs:GetQueryResults`.
{{< /callout >}}

## Instancias EC2

### Instancias Detenidas por Mucho Tiempo
**AWS Doctor** identifica instancias que han estado en estado `stopped` por **más de 30 días**.
- **Motivo**: Aunque no pagas por CPU/RAM cuando está detenida, sigues pagando por los volúmenes raíz EBS adjuntos y cualquier almacenamiento persistente.
- **Acción**: Terminar o hacer un snapshot de los datos y eliminarlos.

{{< callout type="tip" >}}
Puedes personalizar el umbral de 30 días utilizando el flag `--ec2-stopped-days`.
{{< /callout >}}

### Instancias Reservadas (RI) a Punto de Expirar
Busca RIs activas programadas para expirar en los **próximos 30 días** o que hayan expirado en los **últimos 30 días**.
- **Motivo**: Las RIs expiradas vuelven a los costosos precios On-Demand sin previo aviso.
- **Acción**: Revisar el uso y renovar o migrar a Savings Plans.

{{< callout type="tip" >}}
Puedes personalizar el umbral de advertencia de 30 días utilizando el flag `--ec2-ri-expiring-days`.
{{< /callout >}}

### Instancias en Ejecución Inactivas
Encuentra instancias en estado `running` cuya utilización promedio de CPU se mantuvo por debajo del **5%** y cuya combinación de NetworkIn + NetworkOut promedió menos de **5 MB/día** durante los últimos **14 días**.
- **Motivo**: Entornos de desarrollo olvidados, workers abandonados y cargas de trabajo sobre-dimensionadas siguen facturando por cómputo, almacenamiento y EIPs asociadas sin aportar valor.
- **Acción**: Detener la instancia por unos días para verificar que nadie la eche de menos, luego reducir su tamaño o terminarla por completo.

{{< callout type="tip" >}}
Puedes ajustar la sensibilidad de esta verificación utilizando los flags `--ec2-idle-days`, `--ec2-idle-cpu-percent`, y `--ec2-idle-network-bytes`.
{{< /callout >}}

---

## AWS Lambda

### Memoria Sobre-Provisionada
Busca funciones Lambda donde el pico de utilización de memoria es significativamente menor que la asignación configurada (umbral predeterminado: **10%**).
- **Razón**: El precio de Lambda es directamente proporcional a la memoria asignada. Asignar 10GB a una función que usa 200MB desperdicia ~98% del costo.
- **Acción**: Ajustar el tamaño de la memoria de la función según las recomendaciones.
- **Motor de Recomendación**: Sugiere configurar la memoria al **doble del pico observado** (con un mínimo de 128 MB).

{{< callout type="tip" >}}
Puedes ajustar la sensibilidad de esta verificación utilizando el flag `--lambda-memory-threshold` (por ejemplo, `--lambda-memory-threshold 20` para marcar funciones que utilicen menos del 20%). También puedes ajustar la ventana de análisis con el flag `--lambda-lookback-days`.
{{< /callout >}}

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

### Snapshots y AMIs Obsoletas
Marca AMIs y snapshots que tienen **más de 90 días** de antigüedad y no están asociados a ninguna instancia en ejecución o detenida.
- **Motivo**: Imágenes base desactualizadas y copias de seguridad que probablemente no se han tocado en un trimestre.
- **Acción**: Dar de baja las AMIs y eliminar los snapshots para ahorrar costos de almacenamiento.

{{< callout type="tip" >}}
Puedes personalizar el umbral de 90 días de antigüedad utilizando los flags `--ec2-ami-stale-days` y `--ec2-snapshot-stale-days`.
{{< /callout >}}

---

## Acceso y Seguridad

### Key Pairs sin Usar
Identifica los Key Pairs de EC2 que no están asociados con ninguna instancia en ejecución o detenida.
- **Razón**: Reduce el desorden administrativo y los posibles riesgos de seguridad de llaves antiguas.
- **Acción**: Eliminar las llaves sin usar desde la consola/CLI.
