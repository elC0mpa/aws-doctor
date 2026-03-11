---
title: "Almacenamiento y Logs"
description: "Optimice los costos de almacenamiento de S3 y CloudWatch identificando buckets sin políticas de ciclo de vida y grupos de logs sin retención."
weight: 20
---

Optimice sus costos de S3 y CloudWatch asegurando una gestión adecuada del ciclo de vida de los datos y limpiando el desperdicio oculto.

{{< callout type="info" >}}
**Permisos Requeridos**: `s3:ListAllMyBuckets`, `s3:GetLifecycleConfiguration`, `s3:ListBucketMultipartUploads`, `logs:DescribeLogGroups`.
{{< /callout >}}

## Auditoría de Políticas de Ciclo de Vida de S3

**AWS Doctor** escanea cada bucket en su cuenta para verificar si existe una **Configuración de Ciclo de Vida** activa.

### Por qué es importante
Sin una política de ciclo de vida, los datos permanecen en el nivel de almacenamiento Standard (el más caro) para siempre, a menos que se muevan manualmente. Una política puede automatizar:
- La transición de logs antiguos a **IA** (Acceso Infrecuente) o **Glacier**.
- La eliminación automática de datos temporales de trabajo.
- La eliminación de versiones antiguas de objetos.

{{< callout type="warning" >}}
Los buckets sin políticas de ciclo de vida representan un "suelo de costos" que solo crecerá con el tiempo.
{{< /callout >}}

---

## Cargas Multipartes Incompletas en S3

Identifica los buckets que tienen cargas multipartes abandonadas.

### ¿Qué son las Cargas Multipartes?
Cuando carga un archivo grande en S3, este se divide en partes. Si la carga se interrumpe o falla, esas partes permanecen en el bucket ocultas de la vista estándar de objetos.

### El Problema
- **Facturación Oculta**: Se le cobra por el almacenamiento utilizado por estas partes incompletas.
- **Invisibilidad**: No aparecen en `ls` ni en las vistas estándar de la consola.

**AWS Doctor** cuenta estas partes ocultas para que pueda tomar medidas.

### Solución
Agregue una regla de ciclo de vida a su bucket para **"AbortIncompleteMultipartUpload"** después de 7 días.

---

## Retención de Logs en CloudWatch

**AWS Doctor** escanea todos sus **Grupos de Logs de CloudWatch** para identificar aquellos sin política de retención establecida (**"Never Expire"**).

### Por qué es importante
Por defecto, los Grupos de Logs de CloudWatch se crean con un periodo de retención indefinido.
- **Costos Compuestos**: A medida que su aplicación se ejecuta, los logs se acumulan y su factura mensual crece linealmente.
- **Deuda de Almacenamiento**: Muchos desarrolladores crean grupos de logs para servicios temporales o pruebas y nunca establecen una política de limpieza.

### El Problema
- **Fuga de Costos**: Usted paga por cada GB almacenado en CloudWatch. Los logs antiguos e irrelevantes pueden representar una parte significativa de su factura mensual de AWS.
- **Cumplimiento**: Mantener logs indefinidamente podría violar las regulaciones de privacidad de datos como GDPR o SOC2.

### Solución
Establezca un **Periodo de Retención** (ej., 30, 90, o 365 días) para cada Grupo de Logs basándose en sus necesidades de negocio y cumplimiento.
