---
title: "Identidad y Seguridad"
description: "Detecta usuarios IAM sin uso y cuentas Root sin MFA."
weight: 80
type: docs
prev: /docs/waste-detection/configuration
next: /docs/reporting
---

Identity and Access Management (IAM) es la columna vertebral de seguridad de su entorno de AWS. Los usuarios inactivos y las cuentas raíz desprotegidas suponen riesgos de seguridad importantes.

{{< callout type="info" >}}
**Permisos Requeridos**: `iam:GetAccountSummary`, `iam:ListUsers`, `iam:ListAccessKeys`, `iam:GetAccessKeyLastUsed`.
{{< /callout >}}

**AWS Doctor** identifica usuarios de IAM inactivos y comprueba si el usuario raíz de su cuenta de AWS carece de autenticación multifactor (MFA).

## Residuos de IAM y Seguridad

### Lógica de Detección
Un recurso se marca como un riesgo de seguridad o desperdicio si cumple los siguientes criterios:

1. **Cuenta Raíz Desprotegida**: La cuenta raíz no tiene MFA virtual habilitado.
2. **Usuarios IAM Inactivos**: Un usuario de IAM cuya contraseña nunca se ha utilizado (o ha estado inactiva más allá del umbral de inactividad) Y tiene todas las claves de acceso inactivas o sin usar más allá del umbral de inactividad (por defecto: **90 días**).

{{< callout type="tip" >}}
Puede ajustar el umbral de inactividad para usuarios de IAM usando el flag `--iam-idle-days` (ej., `--iam-idle-days 60` para marcar usuarios inactivos por más de 60 días).
{{< /callout >}}

## Cómo Ejecutar
Para ejecutar la detección de residuos de IAM individualmente:

```bash
aws-doctor waste iam
```

## Solución
Si un usuario de IAM o una cuenta raíz es marcado:
1. **Cuenta Raíz**: Inicie sesión como usuario raíz inmediatamente y configure un dispositivo MFA por hardware o virtual.
2. **Usuarios IAM**: Desactive el acceso a la consola o elimine las claves de acceso que ya no se necesiten. Si el usuario es un empleado que se fue, elimine el usuario de IAM por completo.
