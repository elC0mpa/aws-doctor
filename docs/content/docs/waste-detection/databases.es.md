---
title: "Bases de Datos (RDS)"
description: "Audite instancias RDS y snapshots por desperdicio. Identifique bases de datos inactivas y snapshots manuales antiguos para reducir costos."
weight: 15
---

Audite su infraestructura de RDS para eliminar costos de bases de datos abandonadas o subutilizadas.

{{< callout type="info" >}}
**Permisos Requeridos**: `rds:DescribeDBInstances`, `rds:DescribeDBSnapshots`, `cloudwatch:GetMetricStatistics`.
{{< /callout >}}

## Instancias RDS

### Instancias Detenidas
**AWS Doctor** identifica instancias RDS que están en estado `stopped`.
- **Razón**: Mientras los costos de cómputo están pausados, se le sigue facturando por el **almacenamiento asignado** y cualquier IOPS o configuración multi-AZ asociada.
- **Acción**: Si la instancia ya no es necesaria, considere eliminarla después de tomar un snapshot final.

### Instancias RDS Inactivas
Encuentra instancias que están en estado `available` pero que han tenido **cero conexiones a la base de datos** durante los últimos **7 días**.
- **Razón**: Estas instancias están plenamente operativas y facturando por cómputo y almacenamiento, pero no están siendo utilizadas por ninguna aplicación.
- **Acción**: Detenga la instancia si solo es necesaria ocasionalmente, o termínela si ya no se usa.

---

## RDS Snapshots

### Snapshots Manuales Antiguos
Identifica snapshots de RDS manuales que tienen **más de 30 días** de antigüedad.
- **Razón**: A diferencia de los snapshots automáticos que se eliminan según una política de retención, los snapshots manuales persisten indefinidamente hasta que se eliminan manualmente, incurriendo en costos de almacenamiento continuos.
- **Acción**: Revise los snapshots antiguos y elimine aquellos que ya no sean necesarios por cumplimiento o recuperación ante desastres.
