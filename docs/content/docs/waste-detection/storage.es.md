---
title: "Almacenamiento S3"
description: "Optimice los costos de almacenamiento de S3 identificando buckets sin políticas de ciclo de vida y detectando cargas multipartes incompletas ocultas."
weight: 20
---

Optimice sus costos de S3 asegurando una gestión adecuada del ciclo de vida de los datos y limpiando el desperdicio oculto.

{{< callout type="info" >}}
**Permisos Requeridos**: `s3:ListAllMyBuckets`, `s3:GetLifecycleConfiguration`, `s3:ListBucketMultipartUploads`.
{{< /callout >}}

## Auditoría de Políticas de Ciclo de Vida

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

## Cargas Multipartes Incompletas

Identifica los buckets que tienen cargas multipartes abandonadas.

### ¿Qué son las Cargas Multipartes?
Cuando carga un archivo grande en S3, este se divide en partes. Si la carga se interrumpe o falla, esas partes permanecen en el bucket ocultas de la vista estándar de objetos.

### El Problema
- **Facturación Oculta**: Se le cobra por el almacenamiento utilizado por estas partes incompletas.
- **Invisibilidad**: No aparecen en `ls` ni en las vistas estándar de la consola.

**AWS Doctor** cuenta estas partes ocultas para que pueda tomar medidas.

### Solución
Agregue una regla de ciclo de vida a su bucket para **"AbortIncompleteMultipartUpload"** después de 7 días.
