---
title: "Configuración y Secretos"
description: "Detecte secretos de AWS Secrets Manager no utilizados que están facturando pero ya no se acceden."
weight: 70
type: docs
prev: /docs/waste-detection/machine-learning
next: /docs/reporting
---

Los secretos almacenados en **AWS Secrets Manager** cuestan **$0.40 por secreto al mes**, independientemente de si se utilizan activamente. Con el tiempo, los secretos creados para aplicaciones desmanteladas, credenciales rotadas que nunca se limpiaron o pruebas puntuales pueden acumularse en cargos recurrentes significativos.

{{< callout type="info" >}}
**Permisos Requeridos**: `secretsmanager:ListSecrets`.
{{< /callout >}}

**AWS Doctor** identifica los secretos que no han sido accedidos dentro de una ventana de tiempo configurable para que pueda limpiarlos.

## Secrets Manager

### Lógica de Detección
Un secreto se marca como **no utilizado** si cumple con cualquiera de los siguientes criterios:
1. **Nunca accedido**: El campo `LastAccessedDate` es nulo — el secreto nunca ha sido recuperado desde su creación.
2. **Obsoleto**: La `LastAccessedDate` es anterior al umbral de inactividad configurado (por defecto: **90 días**).

{{< callout type="tip" >}}
Puede ajustar el umbral de inactividad utilizando el flag `--secrets-idle-days` (por ejemplo, `--secrets-idle-days 60` para marcar secretos no accedidos en los últimos 60 días).
{{< /callout >}}

{{< callout type="info" >}}
Los secretos réplica (donde `PrimaryRegion` difiere de la región actual) se omiten automáticamente para evitar el doble conteo en configuraciones multi-región.
{{< /callout >}}

## Cómo Ejecutar
Para ejecutar la detección de desperdicio de Secrets Manager de forma individual:

```bash
aws-doctor waste secrets-manager
```

## Remediación
Si un secreto se marca como no utilizado:
1. **Verificar**: Confirme que el secreto ya no es referenciado por ninguna aplicación, función Lambda o pipeline de CI/CD.
2. **Eliminar**: Elimine el secreto a través de la consola de AWS o la CLI. Secrets Manager soporta una ventana de recuperación (por defecto 30 días), por lo que la eliminación es reversible.
3. **Rotar**: Si el secreto aún es necesario pero simplemente fue olvidado, considere habilitar la rotación automática para mantenerlo actualizado.
