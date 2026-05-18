---
title: "Guía de Uso"
weight: 20
type: docs
prev: /docs/getting-started
next: /docs/cost-analytics
---

Aprenda a controlar **AWS Doctor** utilizando subcomandos, flags y perfiles de configuración.

## Subcomandos

Estos son los flujos de trabajo principales de la herramienta.

| Subcomando | Descripción |
| :--- | :--- |
| `cost` | Ejecutar analítica de costos comparativa (Mes actual vs. Mes anterior). |
| `waste` | Ejecutar el motor de detección de desperdicio. |
| `trend` | Generar un informe de tendencia de costos de 6 meses (opcionalmente filtrado por servicio). |
| `report` | Generar reportes PDF profesionales de costos, desperdicio o tendencias. |
| `update` | Actualizar la herramienta a la última versión. |
| `version` | Mostrar información de versión y compilación. |
| `help` | Mostrar ayuda para cualquier subcomando. |

## Flags Globales

Estos flags se pueden usar con cualquier subcomando (incluyendo el análisis de costos por defecto).

| Flag | Por Defecto | Descripción |
| :--- | :--- | :--- |
| `--region` | `~/.aws/config` | Sobrescribir la región de AWS de destino. |
| `--profile` | `default` | Especificar qué perfil de AWS utilizar. |
| `--output` | `table` | Formato de salida: `table`, `json` o `csv`. |

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
aws-doctor cost --profile prod-account
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
aws-doctor update
```
Esto buscará el último lanzamiento en GitHub, descargará el binario para su plataforma y reemplazará el existente.

---

## Rendimiento y Caché

Para garantizar una experiencia rápida y fluida, **AWS Doctor** implementa un servicio de caché local. Esto reduce las llamadas de red redundantes a las APIs externas, mejorando el rendimiento y evitando los límites de velocidad (rate limits).

### Cómo funciona
La caché se aplica actualmente a:
- **Verificaciones de versión**: La comprobación en segundo plano de nuevas versiones se guarda en caché durante **4 horas**.

### Expansión Futura
El mecanismo de caché está diseñado para ser extensible y pronto soportará:
- **Datos de Precios**: Tarifas específicas por región de la API de Precios de AWS.
- **Metadatos de Recursos**: Almacenamiento temporal de respuestas pesadas de la API.

La caché se almacena localmente en el directorio de caché de usuario específico de su SO (por ejemplo, `~/.cache/aws-doctor/` en Linux) y nunca contiene credenciales sensibles de AWS ni datos del usuario.
