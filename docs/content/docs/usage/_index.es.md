---
title: "Guía de Uso"
weight: 20
type: docs
prev: /docs/getting-started
next: /docs/cost-analytics
---

Aprenda a controlar **AWS Doctor** utilizando flags y perfiles de configuración.

## Flags de la CLI

| Flag | Por Defecto | Descripción |
| :--- | :--- | :--- |
| `--region` | `~/.aws/config` | Sobrescribir la región de AWS de destino. |
| `--profile` | `default` | Especificar qué perfil de AWS utilizar. |
| `--waste` | `false` | Ejecutar el motor de detección de desperdicio. |
| `--trend` | `false` | Generar un informe de tendencia de costos de 6 meses. |
| `--output` | `table` | Formato de salida: `table` o `json`. |
| `--update` | `false` | Actualizar la herramienta a la última versión. |
| `--version` | `false` | Mostrar información de versión y compilación. |

---

## Selección de Destino

### Selección de Región
Si no se proporciona el flag `--region`, la herramienta intenta encontrar una región en este orden:
1. Variable de entorno `AWS_REGION`.
2. Variable de entorno `AWS_DEFAULT_REGION`.
3. El campo `region` en su perfil activo dentro de `~/.aws/config`.

### Configuración de Perfil
Para ejecutar auditorías contra una cuenta o rol específico definido en su configuración de AWS:

```bash
aws-doctor --waste --profile prod-account
```

---

## Soporte para MFA

**AWS Doctor** tiene soporte nativo para la Autenticación de Múltiples Factores. Si su perfil utiliza `assume_role` con un `mfa_serial`, la herramienta lo detectará y le solicitará su código de token de forma segura en la terminal.

```text
Enter MFA code for arn:aws:iam::123456789012:mfa/user: ******
```

{{< callout type="info" >}}
La sesión del rol asumido es gestionada por la herramienta. No es necesario ejecutar manualmente `aws sts get-session-token`.
{{< /callout >}}

---

## Actualizaciones Automáticas

Mantenga su motor de diagnóstico actualizado con un solo comando:

```bash
aws-doctor --update
```
Esto buscará el último lanzamiento en GitHub, descargará el binario para su plataforma y reemplazará el existente.
